[![status][ci-status-badge]][ci-status]
[![PkgGoDev][pkg-go-dev-badge]][pkg-go-dev]

# injecuet

injecuet (ɪndʒékjuːt) = inject + [CUE][cuelang]

The injecuet injects concrete values and emits new CUE document.

You can pass emitted new CUE document as `cue eval` or `cue export` argument to generate unified CUE document or other format such as JSON or YAML.
(See [CUE integrations][cue-integrations])

Current supported injection sources:

- environment variables
- Terraform state

## Concept

The injecuet neither have type conversion mechanism nor data validation mechanism by design.

The CUE has [Types are values][types-are-values] concept, so you already get data validation mechanism from CUE.

For example, if you expected the injected values must not be empty, you just write `!= ""` restriction and `cue eval` fails if injected value is empty.

You can also use [CUE modules][cue-modules] such as [strconv][cue-strconv] to parse injected string as integers.

## Synopsis

### Environment variables

```sh
cat src.cue
# name: string @inject(env,name=USER_NAME)

env USER_NAME=aereal injecuet ./src.cue
# name: "aereal" @inject(env,name=USER_NAME)
```

```sh
cat complex.cue
# import "strconv"
# 
# #varAge: string @inject(env,name=AGE)
# age: strconv.Atoi(#varAge)

env AGE=17 injecuet -output ./out.gen.cue ./src.cue

cat out.gen.cue
# #varAge: "123"
# age:     123

cue export ./out.gen.cue
# {
#     "age": 123
# }
```

You can use injecuet with [ssmwrap][ssmwrap]:

```sh
cat src.cue
# secretKey: string @inject(env,name=APP_SECRET_KEY)

ssmwrap -env-prefix APP_ -recursive injecuet ./src.cue
# secretKey: "<value from SSM parameter store>" @inject(env=APP_SECRET_KEY)
```

### Terraform state

You can pass file path or URL to `stateURL`.
Supported URL formats are described on [tfstate-lookup][].

See examples.

## Installation

```sh
go get github.com/aereal/injecuet/cmd/injecuet
```

## Library usage

The injecuet also provides library interface.

See [pkg.go.dev][pkg-go-dev]

## License

See LICENSE file.

[pkg-go-dev]: https://pkg.go.dev/github.com/aereal/injecuet
[pkg-go-dev-badge]: https://pkg.go.dev/badge/aereal/injecuet
[ci-status-badge]: https://github.com/aereal/injecuet/workflows/CI/badge.svg?branch=main
[ci-status]: https://github.com/aereal/injecuet/actions/workflows/CI
[cuelang]: https://cuelang.org/
[cue-integrations]: https://cuelang.org/docs/integrations/
[types-are-values]: https://cuelang.org/docs/concepts/logic/#types-are-values
[cue-modules]: https://cuelang.org/docs/concepts/packages/
[cue-strconv]: https://pkg.go.dev/cuelang.org/go@v0.4.0/pkg/strconv
[ssmwrap]: https://github.com/handlename/ssmwrap
[tfstate-lookup]: https://github.com/fujiwara/tfstate-lookup
