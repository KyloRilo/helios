variable "consul_count" {
  default = 3
}

variable "leader_count" {
  default = 3
}

locals {
  consul_names = [for i in range(var.consul_count) : format("consul-server-%d", i + 1)]
  leader_names = [for i in range(var.leader_count) : format("leader-server-%d", i + 1)]
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
        "./build/consul/server${each.value}_config.json" = "/consul/config/config.json"
      }
      ports = {
        "850${each.value}" = "8500"
        "5${each.value}"   = "53"
      }
    }
  }

  service "consul-client" {
    image   = "hashicorp/consul:latest"
    command = "agent -node=consul-client -config-dir=/consul/config -retry-join=consul-server-1 -retry-join=consul-server-2 -retry-join=consul-server-3"
    volumes = {
      "./build/consul/client_config.json" = "/consul/config/config.json"
    }
    depends_on = local.consul_names
  }

  dynamic "service" {
    for_each = range(var.leader_count)
    labels = [local.leader_names[each.value]]
    content {
      build {
        context    = "."
        dockerfile = "build/docker/leader.Dockerfile"
      }

      hostname = "helios"
      ports = {
        "633${each.value}" = "6330"
      }
      environment = {
        "CONSUL_HTTP_ADDR" = "${local.consul_names[0]}:8500"
        "NODE_ID"          = each.value
        "NODE_COUNT"       = var.leader_count
        "HELIOS_HOST"      = "helios"
        "HELIOS_PORT"      = "6330"
      }

      depends_on = local.consul_names
    }
  }

  service "worker" {
    build {
      context    = "."
      dockerfile = "build/docker/worker.Dockerfile"
    }

    environment = {
      "CONSUL_HTTP_ADDR" = "${local.consul_names[0]}:8500"
      "NODE_ID"          = "2"
      "HELIOS_HOST"      = "helios"
      "HELIOS_PORT"      = "6330"
    }

    depends_on = concat(local.consul_names, local.leader_names)
  }
}
