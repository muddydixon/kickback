Kickback - execute process by api request
----

## Install

```zsh
% go get github.com/muddydixon/kickback
```

## USAGE

* Write settings by yml

```yml
# kickback settings
log:
  dir: ./log
port: 9201
tasks:
  - name: some process
    path: /api/some
    method: POST
    procs:
      - ls {{.DIR}}
      - ls {{.OTHER_DIR}}
  - name: other process
    path: /api/other
    method: get
    procs:
      - ps {{.OPT}}
```

* Then start kickback

```zsh
% kickback --conf ${config}
```

```zsh
% curl -X POST -d dir=/tmp -d OTHER_DIR=/home/sample http://localhost:9201/api/some
% curl -X GET http://localhost:9201/api/other?opt=aux
```
