package eval

import (
	"fmt"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

type Variable struct {
	Name        string         `hcl:"name,label"`
	Type        *hcl.Attribute `hcl:"type,optional"`
	Default     *hcl.Attribute `hcl:"default,optional"`
	Description string         `hcl:"description,optional"`
}

type variableBlock struct {
	Variables []Variable `hcl:"variable,block"`
	Remain    hcl.Body   `hcl:",remain"`
}

type localsBlock struct {
	Locals []localsEntry `hcl:"locals,block"`
	Remain hcl.Body      `hcl:",remain"`
}

type localsEntry struct {
	Attrs hcl.Body `hcl:",remain"`
}

func Functions() map[string]function.Function {
	return map[string]function.Function{
		"format":    stdlib.FormatFunc,
		"join":      stdlib.JoinFunc,
		"upper":     stdlib.UpperFunc,
		"lower":     stdlib.LowerFunc,
		"replace":   stdlib.ReplaceFunc,
		"split":     stdlib.SplitFunc,
		"trimspace": stdlib.TrimSpaceFunc,
		"substr":    stdlib.SubstrFunc,
		"strlen":    stdlib.StrlenFunc,
		"concat":    stdlib.ConcatFunc,
		"coalesce":  stdlib.CoalesceFunc,
		"lookup":    lookupFunc,
		"range":     rangeFunc,
	}
}

var lookupFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "map", Type: cty.DynamicPseudoType},
		{Name: "key", Type: cty.String},
		{Name: "default", Type: cty.DynamicPseudoType},
	},
	Type: function.StaticReturnType(cty.DynamicPseudoType),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		mapVal := args[0]
		key := args[1].AsString()
		def := args[2]

		if !mapVal.Type().IsObjectType() && !mapVal.Type().IsMapType() {
			return def, nil
		}

		if mapVal.Type().IsObjectType() {
			if mapVal.Type().HasAttribute(key) {
				return mapVal.GetAttr(key), nil
			}
			return def, nil
		}

		idx := cty.StringVal(key)
		if mapVal.HasIndex(idx).True() {
			return mapVal.Index(idx), nil
		}
		return def, nil
	},
})

var rangeFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "count", Type: cty.Number},
	},
	Type: function.StaticReturnType(cty.List(cty.Number)),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		n, _ := args[0].AsBigFloat().Int64()
		if n <= 0 {
			return cty.ListValEmpty(cty.Number), nil
		}
		vals := make([]cty.Value, n)
		for i := int64(0); i < n; i++ {
			vals[i] = cty.NumberIntVal(i)
		}
		return cty.ListVal(vals), nil
	},
})

func BuildEvalContext(src []byte, filename string) (*hcl.EvalContext, hcl.Body, error) {
	file, diags := hclsyntax.ParseConfig(src, filename, hcl.InitialPos)
	if diags.HasErrors() {
		return nil, nil, fmt.Errorf("parse error: %s", diags.Error())
	}

	vars, remainAfterVars, err := extractVariables(file.Body)
	if err != nil {
		return nil, nil, err
	}

	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
		Functions: Functions(),
	}

	varNamespace := map[string]cty.Value{}
	for _, v := range vars {
		val, err := resolveVariable(v, ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("variable %q: %s", v.Name, err)
		}
		varNamespace[v.Name] = val
	}
	ctx.Variables["var"] = cty.ObjectVal(safeObject(varNamespace))

	locals, remainAfterLocals, err := extractLocals(remainAfterVars, ctx)
	if err != nil {
		return nil, nil, err
	}
	ctx.Variables["local"] = cty.ObjectVal(safeObject(locals))

	if err := expandDynamicBlocks(remainAfterLocals, ctx); err != nil {
		return nil, nil, fmt.Errorf("expanding dynamic blocks: %s", err)
	}

	resources := extractResourceRefs(remainAfterLocals)
	for nsName, ns := range resources {
		ctx.Variables[nsName] = cty.ObjectVal(safeObject(ns))
	}

	return ctx, remainAfterLocals, nil
}

func BuildFileEvalContext(path string, src []byte) (*hcl.EvalContext, error) {
	file, diags := hclsyntax.ParseConfig(src, path, hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf("parse error: %s", diags.Error())
	}

	vars, remainAfterVars, err := extractVariables(file.Body)
	if err != nil {
		return nil, err
	}

	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
		Functions: Functions(),
	}

	varNamespace := map[string]cty.Value{}
	for _, v := range vars {
		val, err := resolveVariable(v, ctx)
		if err != nil {
			return nil, fmt.Errorf("variable %q: %s", v.Name, err)
		}
		varNamespace[v.Name] = val
	}
	ctx.Variables["var"] = cty.ObjectVal(safeObject(varNamespace))

	locals, remainAfterLocals, err := extractLocals(remainAfterVars, ctx)
	if err != nil {
		return nil, err
	}
	ctx.Variables["local"] = cty.ObjectVal(safeObject(locals))

	if err := expandDynamicBlocks(remainAfterLocals, ctx); err != nil {
		return nil, fmt.Errorf("expanding dynamic blocks: %s", err)
	}

	resources := extractResourceRefs(remainAfterLocals)
	for nsName, ns := range resources {
		ctx.Variables[nsName] = cty.ObjectVal(safeObject(ns))
	}

	return ctx, nil
}

func extractVariables(body hcl.Body) ([]Variable, hcl.Body, error) {
	var vb variableBlock
	diags := gohcl.DecodeBody(body, nil, &vb)
	if diags.HasErrors() {
		return nil, nil, fmt.Errorf("decoding variables: %s", diags.Error())
	}
	return vb.Variables, vb.Remain, nil
}

func resolveVariable(v Variable, ctx *hcl.EvalContext) (cty.Value, error) {
	if v.Default == nil {
		return cty.NilVal, fmt.Errorf("variable %q has no default value and no value was provided", v.Name)
	}

	val, diags := v.Default.Expr.Value(ctx)
	if diags.HasErrors() {
		return cty.NilVal, fmt.Errorf("evaluating default for %q: %s", v.Name, diags.Error())
	}
	return val, nil
}

func extractLocals(body hcl.Body, ctx *hcl.EvalContext) (map[string]cty.Value, hcl.Body, error) {
	var lb localsBlock
	diags := gohcl.DecodeBody(body, nil, &lb)
	if diags.HasErrors() {
		return nil, nil, fmt.Errorf("decoding locals: %s", diags.Error())
	}

	allAttrs := map[string]*hcl.Attribute{}
	for _, entry := range lb.Locals {
		attrs, diags := entry.Attrs.JustAttributes()
		if diags.HasErrors() {
			return nil, nil, fmt.Errorf("reading locals attributes: %s", diags.Error())
		}
		for name, attr := range attrs {
			allAttrs[name] = attr
		}
	}

	ordered := sortLocalsByDeps(allAttrs)

	locals := map[string]cty.Value{}
	for _, name := range ordered {
		attr := allAttrs[name]
		localCtx := ctx.NewChild()
		localCtx.Variables = map[string]cty.Value{
			"var":   ctx.Variables["var"],
			"local": cty.ObjectVal(safeObject(locals)),
		}
		localCtx.Functions = ctx.Functions

		val, diags := attr.Expr.Value(localCtx)
		if diags.HasErrors() {
			return nil, nil, fmt.Errorf("evaluating local %q: %s", name, diags.Error())
		}
		locals[name] = val
	}

	return locals, lb.Remain, nil
}

func expandDynamicBlocks(body hcl.Body, ctx *hcl.EvalContext) error {
	syntaxBody, ok := body.(*hclsyntax.Body)
	if !ok {
		return nil
	}
	return expandDynamic(syntaxBody, ctx)
}

func expandDynamic(body *hclsyntax.Body, ctx *hcl.EvalContext) error {
	var newBlocks []*hclsyntax.Block

	for _, block := range body.Blocks {
		if block.Type != "dynamic" || len(block.Labels) == 0 {
			if block.Body != nil {
				if err := expandDynamic(block.Body, ctx); err != nil {
					return err
				}
			}
			newBlocks = append(newBlocks, block)
			continue
		}

		expanded, err := expandSingleDynamic(block, ctx)
		if err != nil {
			return err
		}
		newBlocks = append(newBlocks, expanded...)
	}

	body.Blocks = newBlocks
	return nil
}

func expandSingleDynamic(block *hclsyntax.Block, ctx *hcl.EvalContext) ([]*hclsyntax.Block, error) {
	targetType := block.Labels[0]

	forEachAttr, ok := block.Body.Attributes["for_each"]
	if !ok {
		return nil, fmt.Errorf("dynamic %q block missing for_each attribute", targetType)
	}
	forEachVal, diags := forEachAttr.Expr.Value(ctx)
	if diags.HasErrors() {
		return nil, fmt.Errorf("evaluating for_each in dynamic %q: %s", targetType, diags.Error())
	}

	var contentBlock *hclsyntax.Block
	for _, b := range block.Body.Blocks {
		if b.Type == "content" {
			contentBlock = b
			break
		}
	}
	if contentBlock == nil {
		return nil, fmt.Errorf("dynamic %q block missing content block", targetType)
	}

	labelsAttr := block.Body.Attributes["labels"]

	keys, values := iterateForEach(forEachVal)
	var result []*hclsyntax.Block

	for i, key := range keys {
		val := values[i]

		iterCtx := ctx.NewChild()
		iterCtx.Variables = map[string]cty.Value{
			"each": cty.ObjectVal(map[string]cty.Value{
				"key":   key,
				"value": val,
			}),
		}
		iterCtx.Functions = ctx.Functions

		var labels []string
		if labelsAttr != nil {
			labelsVal, diags := labelsAttr.Expr.Value(iterCtx)
			if diags.HasErrors() {
				return nil, fmt.Errorf("evaluating labels in dynamic %q: %s", targetType, diags.Error())
			}
			for it := labelsVal.ElementIterator(); it.Next(); {
				_, lv := it.Element()
				labels = append(labels, lv.AsString())
			}
		} else if val.Type() == cty.String {
			labels = []string{val.AsString()}
		} else {
			n, _ := key.AsBigFloat().Int64()
			labels = []string{fmt.Sprintf("%s-%d", targetType, n)}
		}

		newBody := cloneBodyForIteration(contentBlock.Body, iterCtx)

		expanded := &hclsyntax.Block{
			Type:   targetType,
			Labels: labels,
			Body:   newBody,
		}
		result = append(result, expanded)
	}

	return result, nil
}

func iterateForEach(val cty.Value) ([]cty.Value, []cty.Value) {
	var keys, values []cty.Value

	ty := val.Type()
	switch {
	case ty.IsListType() || ty.IsTupleType() || ty.IsSetType():
		i := int64(0)
		for it := val.ElementIterator(); it.Next(); {
			_, v := it.Element()
			keys = append(keys, cty.NumberIntVal(i))
			values = append(values, v)
			i++
		}
	case ty.IsMapType() || ty.IsObjectType():
		for it := val.ElementIterator(); it.Next(); {
			k, v := it.Element()
			keys = append(keys, k)
			values = append(values, v)
		}
	case ty == cty.Number:
		n, _ := val.AsBigFloat().Int64()
		for i := int64(0); i < n; i++ {
			keys = append(keys, cty.NumberIntVal(i))
			values = append(values, cty.NumberIntVal(i))
		}
	}

	return keys, values
}

func cloneBodyForIteration(body *hclsyntax.Body, iterCtx *hcl.EvalContext) *hclsyntax.Body {
	newBody := &hclsyntax.Body{
		Attributes: make(hclsyntax.Attributes, len(body.Attributes)),
		SrcRange:   body.SrcRange,
		EndRange:   body.EndRange,
	}

	for name, attr := range body.Attributes {
		if exprReferencesVar(attr.Expr, "each") {
			val, diags := attr.Expr.Value(iterCtx)
			if !diags.HasErrors() {
				newBody.Attributes[name] = &hclsyntax.Attribute{
					Name:        name,
					Expr:        &hclsyntax.LiteralValueExpr{Val: val, SrcRange: attr.SrcRange},
					SrcRange:    attr.SrcRange,
					NameRange:   attr.NameRange,
					EqualsRange: attr.EqualsRange,
				}
			} else {
				newBody.Attributes[name] = attr
			}
		} else {
			newBody.Attributes[name] = attr
		}
	}

	for _, blk := range body.Blocks {
		newBlk := &hclsyntax.Block{
			Type:      blk.Type,
			Labels:    blk.Labels,
			Body:      cloneBodyForIteration(blk.Body, iterCtx),
			TypeRange: blk.TypeRange,
		}
		newBody.Blocks = append(newBody.Blocks, newBlk)
	}

	return newBody
}

func exprReferencesVar(expr hcl.Expression, name string) bool {
	for _, traversal := range expr.Variables() {
		if traversal.RootName() == name {
			return true
		}
	}
	return false
}

var resourceBlockTypes = []string{"service", "cluster"}

func extractResourceRefs(body hcl.Body) map[string]map[string]cty.Value {
	resources := map[string]map[string]cty.Value{}

	syntaxBody, ok := body.(*hclsyntax.Body)
	if !ok {
		return resources
	}

	collectBlocks(syntaxBody, resources)
	return resources
}

func collectBlocks(body *hclsyntax.Body, resources map[string]map[string]cty.Value) {
	for _, block := range body.Blocks {
		isResource := false
		for _, rt := range resourceBlockTypes {
			if block.Type == rt {
				isResource = true
				break
			}
		}

		if isResource && len(block.Labels) > 0 {
			name := block.Labels[0]
			if resources[block.Type] == nil {
				resources[block.Type] = map[string]cty.Value{}
			}
			resources[block.Type][name] = cty.StringVal(name)
		}

		if block.Body != nil {
			collectBlocks(block.Body, resources)
		}
	}
}

func sortLocalsByDeps(attrs map[string]*hcl.Attribute) []string {
	names := make([]string, 0, len(attrs))
	for name := range attrs {
		names = append(names, name)
	}
	sort.Strings(names)

	deps := map[string][]string{}
	for name, attr := range attrs {
		var localDeps []string
		for _, traversal := range attr.Expr.Variables() {
			if traversal.RootName() == "local" && len(traversal) > 1 {
				if step, ok := traversal[1].(hcl.TraverseAttr); ok {
					localDeps = append(localDeps, step.Name)
				}
			}
		}
		deps[name] = localDeps
	}

	visited := map[string]bool{}
	var ordered []string
	var visit func(string)
	visit = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true
		for _, dep := range deps[name] {
			if _, exists := attrs[dep]; exists {
				visit(dep)
			}
		}
		ordered = append(ordered, name)
	}
	for _, name := range names {
		visit(name)
	}
	return ordered
}

func safeObject(m map[string]cty.Value) map[string]cty.Value {
	if len(m) == 0 {
		return map[string]cty.Value{"_": cty.StringVal("")}
	}
	return m
}
