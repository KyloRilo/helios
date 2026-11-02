variable "consul_count" {
  default = 3
}

variable "leader_count" {
  default = 3
}

locals {
  leader_names = [for i in range(var.leader_count) : format("helios-leader-%d", i + 1)]
  consul_names = [for i in range(var.consul_count) : format("consul-server-%d", i + 1)]
  join_cmd = join(" -retry-join=", local.consul_names)
}

cluster "qa" {
  dynamic "service" {
    for_each = range(var.consul_count)
    labels   = [local.consul_names[each.value]]
    content {
      image    = "hashicorp/consul:latest"
      hostname = local.consul_names[each.value]
      command  = "agent -server -bootstrap-expect=3 -node=${local.consul_names[each.value]}"
      volumes  = {
        "bin/consul/server${each.value + 1}_config.json" = "/consul/config/config.json"
      }
      ports = {
        "850${each.value}" = "8500"
      }
    }
  }

  service "consul-client" {
    image   = "hashicorp/consul:latest"
    command = "agent -node=consul-client -config-dir=/consul/config ${local.join_cmd}"
    volumes = {
      "bin/consul/client_config.json" = "/consul/config/config.json"
    }
    depends_on = local.consul_names
  }

  dynamic "service" {
    for_each = range(var.leader_count)
    labels = [local.leader_names[each.value]]
    content {
      build {
        context    = "."
        dockerfile = "bin/docker/leader.Dockerfile"
      }

      hostname = local.leader_names[each.value]
      ports = {
        "633${each.value}" = "6330"
        "634${each.value}" = "6331"
      }
      volumes = {
        "bin/config/leader-${each.value + 1}.hcl" = "/helios/config/config.hcl"
        "/var/run/docker.sock"                     = "/var/run/docker.sock"
      }
      environment = {
        "CONSUL_HTTP_ADDR" = "${local.consul_names[0]}:8500"
      }

      depends_on = local.consul_names
    }
  }
}
