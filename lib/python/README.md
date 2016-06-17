# Frugal Python

## Using

```bash
pip install frugal
```
or to use with the [Tornado nats-client](https://github.com/nats-io/python-nats)
```bash
pip install frugal[tornado]
```
or preferably add one of the following to requirements.txt
```bash
frugal=={version}
# or for tornado support
frugal[tornado]=={version}
```
## Contributing
1. Make a virutalenv `mkvirtualenv frugal -a /path/to/frugal/lib/python`
2. Install dependecies `make deps`
3. Write code, tests & create a pull requests
    a. Automatically run tests on fail save with `make sniffer`
