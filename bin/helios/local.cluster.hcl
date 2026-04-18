cluster "test-cluster" {
    service "consul" {
        image = "hashicorp/consul:latest"
        command = "agent -dev"
        ports = {
            "8500":"8500", 
            "53":"53"
        }
    }

    service "localstack" {
        image = "localstack/localstack:latest"
        ports = { "4566":"4566" }
        volumes = { "/var/run/docker.sock":"/var/run/docker.sock" }
    }

    service "core" {
        build {
            context = "."
            dockerfile = "build/docker/leader.Dockerfile"
        }
        
        ports = { "6331":"6330" }
        environment = { "CONSUL_HTTP_ADDR":"localhost:8500" }
        depends_on = [ "consul", "localstack" ]
    }
    
    service "worker" {
        build {
            context = "."
            dockerfile = "build/docker/worker.Dockerfile"
        }
        
        environment = { "CONSUL_HTTP_ADDR":"localhost:8500"}
        depends_on = [ "consul" , "core" ]
    }

    service "api" {
        build {
            context = "."
            dockerfile = "build/docker/api.Dockerfile"
        }
        
        ports = { "8080":"8080" }
        environment = { "CONSUL_HTTP_ADDR":"localhost:8500" }
        depends_on = [ "consul", "core" ]
    }
}