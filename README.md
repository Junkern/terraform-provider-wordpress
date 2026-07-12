# Terraform Wordpress Provider

Terraform provider for managing WordPress content through the WordPress REST API.

## Overview

The provider connects to a WordPress site via its REST API and uses application password authentication.

## Supported Resources

- `wordpress_page` - manage WordPress pages
- `wordpress_plugin` - install, activate, and remove WordPress plugins
- `wordpress_post` - manage WordPress posts
- `wordpress_user` - manage WordPress users

## Supported Data Sources

- `wordpress_plugin_info` - read public plugin metadata from the WordPress.org registry by slug
- `wordpress_pages` - read a list of WordPress pages
- `wordpress_plugins` - read a list of installed WordPress plugins
- `wordpress_posts` - read a list of WordPress posts
- `wordpress_users` - read a list of WordPress users


## Using the provider

### Provider configuration

Configure the provider with the following settings:

- `host` - the base URL of the WordPress REST API, such as `http://localhost:8888/wp-json/wp/v2`
- `username` - the WordPress username to authenticate with
- `password` - a WordPress application password

These values can also be supplied via environment variables: `WP_TF_PROVIDER_HOST`, `WP_TF_PROVIDER_USERNAME`, and `WP_TF_PROVIDER_PASSWORD`. The legacy `WORDPRESS_*` names are still accepted.

Example:

```hcl
provider "wordpress" {
	host     = "http://localhost:8888/wp-json/wp/v2"
	username = "admin"
	password = "application-password"
}
```


## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
