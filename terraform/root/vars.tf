variable "helios_role_config" {
    type = map(object({
      role_arn = string
      external_id = string
    }))
}

variable "region" {
    type = string
}