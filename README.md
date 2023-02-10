# gqlgen-depth-limit-extension

This extension is used by [gqlgen](https://github.com/99designs/gqlgen). Default complexity limit logic in gqlgen shares a single limit for complexities and depths. This extension allows explicitly setting the depth limit.

## Usage

```golang
server.Use(depth.FixedDepthLimit(5))
```
