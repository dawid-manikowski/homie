# homie
The simple homelab monitoring tool.

![Screenshot](/screen.png?raw=true "Homie")

## Configure
First fill in all the config info in .envrc file
```bash
cp .envrc.example .envrc
vim .envrc # use your favorite editor
```

## Run
Then just run
```bash
bash -c "go build && source .envrc && ./homie"
```

