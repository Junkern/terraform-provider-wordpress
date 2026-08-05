provider "wordpress" {
  host = "http://localhost:8888/wp-json/wp/v2"

  app_auth {
    username = "username"
    password = "application_password"
  }
}
