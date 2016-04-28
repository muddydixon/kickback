Kickback - execute process by api server
----

## Install

```zsh
% go get github.com/muddydixon/kickback
```

## USAGE

* Write settings by yml

```yml
# kickback tasks
log:
  dir: ./log
  level: debug
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
    method: PUT
    procs:
      - ls
```

* Then start kickback

```zsh
% kickback --taskfile ${taskfile}
```

```zsh
% curl -X POST -d DIR=/tmp -d OTHER_DIR=/home/sample http://localhost:9201/api/some
```
