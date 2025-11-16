cluster "test-cluster" {
    host = "localhost"
    port = 6330
    service "leader" {
        build {
            context = "."
            dockerfile = "build/docker/leader.Dockerfile"
        }
        
        ports = [ "6331:6330" ]
        environment = [
            "CONSUL_HTTP_ADDR=consul-server-1:8500",
            "NODE_ID=1",
            "HELIOS_HOST=helios",
            "HELIOS_PORT=6330",
        ]
    }

    service "consul-1" {
        image = "hashicorp/consul:latest"
        command = "agent -server -bootstrap-expect=3 -node=consul-1"
        volumes = [ "./build/consul/server1_config.json:/consul/config/config.json" ]
    }

    service "consul-2" {
        image = "hashicorp/consul:latest"
        command = "agent -server -bootstrap-expect=3 -node=consul-2"
        volumes = [ "./build/consul/server1_config.json:/consul/config/config.json" ]
    }

    service "consul-3" {
        image = "hashicorp/consul:latest"
        command = "agent -server -bootstrap-expect=3 -node=consul-3"
        volumes = [ "./build/consul/server1_config.json:/consul/config/config.json" ]
    }
}