cluster "integration" {
  service "redis" {
    image   = "redis:7-alpine"
    command = "redis-server"
    ports   = { "6379" = "6379" }
  }

  service "consul" {
    image   = "hashicorp/consul:latest"
    command = "agent -dev -client=0.0.0.0"
    ports   = { "8500" = "8500" }
  }

  service "api" {
    image      = "nginx:alpine"
    ports      = { "8080" = "80" }
    depends_on = ["redis", "consul"]
  }
}
