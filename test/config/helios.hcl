cluster "test-cluster" {
    host = "localhost"
    port = 6330
    service "test-build" {
        build {
            context = "."
            dockerfile = "build/docker/leader.Dockerfile"
        }
        
        ports = [ "6331:6330" ]
        environment = ["ENV_VAR=true"]
        depends_on = [ "test-image-1", "test-image-2", "test-image-3" , "test-client"]
    }

    service "test-image-1" {
        image = ""
        command = ""
        volumes = [ "" ]
    }

    service "test-image-2" {
        image = ""
        command = ""
        volumes = [ "" ]
    }

    service "test-image-3" {
        image = ""
        command = ""
        volumes = [ "" ]
    }

    service "test-client" {
        image = ""
        command = ""
        volumes = [ "" ]
        depends_on = [ "test-image-1", "test-image-2", "test-image-3" ]
    }
}