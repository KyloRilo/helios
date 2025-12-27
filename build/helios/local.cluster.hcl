cluster "test-cluster" {
    host = "localhost"
    port = 6330

    service "consul-1" {
        image = "hashicorp/consul:latest"
        command = "agent -server -bootstrap-expect=3 -node=consul-1"
        volumes = [ "./build/consul/server1_config.json:/consul/config/config.json" ]
        ports = [ "8500:8500", "53:53" ]
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

    service "consul-client" {
        image = "hashicorp/consul:latest"
        command = "agent -node=consul-client -config-dir=/consul/config -retry-join=consul-server-1 -retry-join=consul-server-2 -retry-join=consul-server-3"
        volumes = [ "./build/consul/client_config.json:/consul/config/config.json" ]
        depends_on = [ "consul-1", "consul-2", "consul-3" ]
    }

    service "leader" {
        build {
            context = "."
            dockerfile = "build/docker/leader.Dockerfile"
        }
        
        ports = [ "6331:6330" ]
        environment = [
            "CONSUL_HTTP_ADDR=consul-1:8500",
            "NODE_ID=1",
            "HELIOS_HOST=helios",
            "HELIOS_PORT=6330",
        ]
        
        depends_on = [ "consul-1", "consul-2", "consul-3" ]
    }
}