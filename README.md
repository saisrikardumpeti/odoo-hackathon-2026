# Odoo Hackathon 2026

```bash
# Debian based
curl -1sLf 'https://dl.cloudsmith.io/public/task/task/setup.deb.sh' | sudo -E bash
apt install task

# MacOS
brew install go-task/tap/go-task

# windows
winget install Task.Task
```

other os check out tasks offical docs [here](https://taskfile.dev/docs/installation) 

## installing dependencies

```bash
task setup:backend
task setup:frontend
```

## Dev environment
```bash
task dev
```
