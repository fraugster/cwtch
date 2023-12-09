<h1 align="center">cwtch</h1>
<p align="center">
        <a href="https://github.com/fraugster/cwtch/releases"><img src="https://img.shields.io/github/tag/fraugster/cwtch.svg?label=version&color=brightgreen"></a>
        <a href="https://goreportcard.com/report/github.com/fraugster/cwtch"><img src="https://goreportcard.com/badge/github.com/fraugster/cwtch"></a>
        <a href="https://pkg.go.dev/github.com/fraugster/cwtch"><img src="https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=for-the-badge"></a>
        <a href="https://github.com/fraugster/cwtch/blob/master/LICENSE"><img src="https://img.shields.io/badge/license-Apache--2-blue.svg"></a>
</p>

---

cwtch is a `watch(1)` with configurable highlighting. It implements a subset of the options available in `watch(1)`.

## Getting Started

Install `cwtch` from your command line using `go install`:

```bash
go install github.com/fraugster/cwtch@latest
```

## Configuring highlighting

The default configuration file is `~/.cwtch.yml`. You can provide alternative configuration through the `--config` option.
Alternatively, you can also use multiple configuration files by putting them all in a common configuration directory `~/.cwtch`. This
configuration directory is configurable through the `--config-dir` option. Only one of these configuration files will be
chosen, depending on a logic described below.

Here is an example file of some configuration:

```yaml
cmd_regex: kubectl
highlights:
  - regex: Completed|Succeeded|Bound|Released
    fg: blue, bold
    bg: white
```

`highlights` is a list of items. Each item configures a highlighting regular 
expression. The regular expression is provided in `regex`. Any text matched in 
the command output is highlighted using the provided colours. In this case, the 
foreground colour is set to `blue` and `bold`, while the background is set to 
`white`. If `fg` or `bg` is omitted, the default colour is used. Any colour can 
only be used. Attributes are optional and can be combined with each other.

The optional attribute `cmd_regex` defines a regular expression. If this attribute 
is set, then the highlights associated with it will only be applied if the regular 
expression matches the provided command. This is useful when using multiple 
configuration files: when executing a command, `cwtch` will search through all 
configuration files and will choose the first configuration file (sorted alphanumerically) 
either has no `cmd_regex` set or whose `cmd_regex` matches the command that gets
executed. If `~/.cwtch.yml` is present, it will always be evaluated last.

List of available colours:

* `black`
* `red`
* `green`
* `yellow`
* `blue`
* `magenta`
* `cyan`
* `white`

List of available attributes:

* `bold`
* `underline`
* `reverse`
* `italic`
* `blink`
* `dim`
* `strikethrough`

For more examples, have a look at the `examples/` subdirectory.

## Contributing

If you want to hack on this repository, please read the short [CONTRIBUTING.md](CONTRIBUTING.md)
guide first.

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions, available,
see the [tags on this repository][tags].

## Authors

- **Andreas Krennmair** - *Initial work* [akrennmair](https://github.com/akrennmair)

See also the list of [contributors][contributors] who participated in this project.

## License

This project is licensed under the Apache-2 license - see the [LICENSE](LICENSE) file for details.

[tags]: https://github.com/fraugster/cwtch/tags
[contributors]: https://github.com/fraugster/cwtch/graphs/contributors
