Kickback - execute process by api server
----

## Install

```zsh
% go get github.com/muddydixon/kickback
```

## USAGE

* Write settings by yml

```yml
- name: some process
  path: /api/some
  method: POST
  procs:
    - ls {{.DIR}}
    - ls {{.OTHER_DIR}}
```

* Then start kickback

```zsh
% kickback --taskfile ${taskfile}
```

```zsh
% curl -X POST -d DIR=/tmp -d OTHER_DIR=/home/sample http://localhost:8080/api/some
```
