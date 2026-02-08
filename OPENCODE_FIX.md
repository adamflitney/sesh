# OpenCode + Neovim Integration Fix

## The Problem

The `opencode.nvim` plugin couldn't connect to OpenCode because:
1. OpenCode wasn't started with the `--port` flag (required for server mode)
2. The neovim config was trying to create its own OpenCode instance instead of connecting to the tmux one

## The Solution

### 1. Updated `sesh` to Start OpenCode with Port Flag

**File**: `/Users/adamflitney/dev/sesh/internal/tmux/session.go`

Changed line 69 from:
```go
"opencode ."
```

To:
```go
"opencode --port 0 ."
```

The `--port 0` flag tells OpenCode to:
- Start a server on a random available port
- Enable `opencode.nvim` to connect to it
- Allow neovim to send commands and context

### 2. Updated Neovim Config to Use External OpenCode

**File**: `~/.config/nvim/lua/plugins/opencode.lua`

Changed the provider from:
```lua
provider = {
  enabled = "snacks",
  snacks = { ... }
}
```

To:
```lua
provider = {
  enabled = false,  -- Use external OpenCode from tmux
}
```

This tells `opencode.nvim` to:
- NOT create its own OpenCode instance
- Look for an existing OpenCode server in the current directory
- Connect to the OpenCode running in the tmux window

## How to Apply the Fix

### Step 1: Rebuild sesh
The updated `sesh` tool has been rebuilt and installed:
```bash
cd ~/dev/sesh
go install
```

### Step 2: Kill Existing OpenCode Sessions
```bash
killall opencode
tmux kill-session -t <your-session-name>
```

### Step 3: Start Fresh with sesh
```bash
sesh
# Select your project
```

This will create a new session with OpenCode properly configured.

### Step 4: Verify OpenCode is Running with Port

Run the diagnostic script:
```bash
~/dev/sesh/scripts/check-opencode.sh
```

You should see:
```
✓ OpenCode server listening:
  Port: 127.0.0.1:XXXXX
```

### Step 5: Reload Neovim Config

In neovim (window 1):
```vim
:source ~/.config/nvim/init.lua
```

Or restart neovim.

### Step 6: Test the Integration

1. In neovim, open a file
2. Select some code (visual mode)
3. Press `<leader>oe` (Explain code)
4. Should:
   - Send code to OpenCode
   - Automatically switch to OpenCode tmux window
   - Show explanation

## Diagnostic Script

Use this script anytime to check if OpenCode is properly configured:

```bash
~/dev/sesh/scripts/check-opencode.sh
```

It checks:
- ✓ OpenCode is running
- ✓ OpenCode is listening on a port
- ✓ Running in tmux
- ✓ OpenCode tmux window exists

## How It Works Now

```
┌─────────────────┐
│  Neovim Window  │
│                 │
│  1. Select code │
│  2. Press <l>oe │
└────────┬────────┘
         │
         │ opencode.nvim sends
         │ command via HTTP
         ▼
┌─────────────────┐
│ OpenCode Server │ ← Running with --port flag
│ (tmux window 2) │
│                 │
│ Port: 127.0.0.1 │
│       :XXXXX    │
└─────────────────┘
         │
         │ tmux auto-switch
         ▼
┌─────────────────┐
│ You see the     │
│ OpenCode window │
│ with response   │
└─────────────────┘
```

## Troubleshooting

### "No OpenCode server listening"
- OpenCode wasn't started with `--port` flag
- Solution: Kill and restart with sesh

### "opencode.nvim can't connect"
- Check neovim config has `provider.enabled = false`
- Make sure you're in the same directory where OpenCode started

### "Commands don't switch windows"
- Make sure you're in a tmux session (created by sesh)
- Check tmux has an "opencode" window: `tmux list-windows`

### Still not working?
Run the diagnostic:
```bash
~/dev/sesh/scripts/check-opencode.sh
```

And check the neovim OpenCode plugin status:
```vim
:checkhealth opencode
```

## Files Changed

1. ✅ `/Users/adamflitney/dev/sesh/internal/tmux/session.go` - Added `--port 0` flag
2. ✅ `~/.config/nvim/lua/plugins/opencode.lua` - Disabled internal provider
3. ✅ `~/dev/sesh/scripts/check-opencode.sh` - New diagnostic script

## Next Steps

After applying the fix:
1. Test basic commands (`<leader>oa`, `<leader>oe`)
2. Verify auto-switching works
3. Try the full workflow (select code → explain → read response → return to neovim)
