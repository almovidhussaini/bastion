if you want to run locally

go run ./cmd/daemon
go run ./cmd/bastion
cd  web
npm install
npm run dev

set this to env file
BASTION_DB_DSN=postgres://postgres:newpass123@localhost:5432/bastion?sslmode=disable
DAEMON_URL=http://localhost:9081


if you want to run daemon  on ubundu ser
Setting Up SSH Port Forwarding (Local → Remote) in the window bash
ssh -N -L 9081:127.0.0.1:9081 shah@154.57.209.191

go run ./cmd/bastion
cd  web
npm install
npm run dev



workflow example
1) creating hello world command 
   react app sends a post request to Bastion at /api/v1/commands with {name.descriptuon,scripts,timeout_seconds} Bastion's hander is in cmd/baston/main.go{handleCommands},which decodes the json and call svc.CreateCommands
   core.BastionService in {cmd/internal/core} validates and stores the command in db

2) script execution
    when you run the command from the UI, the fronend POSTs to /api/v1/execute with command_id and node_id
    handleExecute in cmd/bastion/main.go looks up the command and node, then call 
    svc.EcecuteCommand
    svc.ExecuteCommand forwards the script and timeout to the selected node's address, hitting that daemon's /api/v1/exec
    daemon runs the scripts in cmd/daemon/main.go, handleExec received the request and runScript executes bash -lc via exec.CommandContext on the daemon host


