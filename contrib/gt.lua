-- gt.lua
-- Neovim configurations for the 'gt' terminal

local M = {}

-- Helper function to check if running inside 'gt' terminal
-- This is a placeholder; a more robust detection method might be needed.
-- For example, 'gt' could set a specific environment variable like $GT_TERMINAL_VERSION
local function is_gt_terminal()
  -- Simple check based on TERM, assuming 'gt' sets it to 'xterm-256color' or similar for now.
  -- A more specific identifier from 'gt' would be better.
  -- return vim.env.TERM == "gt" or vim.env.GT_TERMINAL == "1"
  return vim.env.TERM == "xterm-256color" -- This is a common default, adjust if 'gt' uses a unique $TERM
end

function M.setup()
  if not is_gt_terminal() then
    -- vim.notify("Not running in 'gt' terminal, 'gt.lua' setup skipped.", vim.log.levels.INFO)
    -- return -- Or, apply some settings anyway if they are harmless
  end

  vim.notify("gt.lua: Setting up for 'gt' terminal...", vim.log.levels.INFO)

  -- 1. Cursor Shape (DECSCUSR)
  -- These rely on 'gt' interpreting these standard escape codes.
  -- \x1b[0q or \x1b[ q - blinking block
  -- \x1b[1q - blinking block (often same as 0)
  -- \x1b[2q - steady block
  -- \x1b[3q - blinking underline
  -- \x1b[4q - steady underline
  -- \x1b[5q - blinking bar (I-beam)
  -- \x1b[6q - steady bar (I-beam)

  vim.opt.guicursor = table.concat({
    "n-v-c:block-Cursor/lCursor",      -- Normal, visual, command-line: block cursor
    "i-ci-ve:ver25-Cursor/lCursor",    -- Insert, command-line insert, visual-explorer: vertical bar (25% width)
    "r-cr-o:hor20-Cursor/lCursor",     -- Replace, command-line replace, operator-pending: horizontal bar (20% height)
    "sm:block-Cursor-blinkwait175-blinkoff150-blinkon175" -- Showmatch: blinking block
  }, ",")

  -- More direct way to try and set cursor shapes using raw escape codes if guicursor doesn't translate perfectly
  -- This is often needed for terminals that don't fully support the guicursor options but do support DECSCUSR.
  local group = vim.api.nvim_create_augroup("GtCursorShape", { clear = true })
  vim.api.nvim_create_autocmd("ModeChanged", {
    group = group,
    pattern = "*",
    callback = function()
      local mode = vim.fn.mode()
      local term_program = vim.v.termresponse -- Check if this is a terminal buffer with a running program
      if term_program ~= nil and term_program ~= "" then
        -- Don't change cursor shape if in a terminal buffer with a running program
        return
      end

      if mode:find("^[nvcV]") then -- Normal, Visual, Command-line, Select modes
        vim.fn.execute('silent! !printf "\\x1b[2 q"', "") -- Steady Block
      elseif mode:find("^[iR]") then -- Insert, Replace modes
        if mode:find("^R") then -- Specifically Replace mode
            vim.fn.execute('silent! !printf "\\x1b[4 q"', "") -- Steady Underline
        else -- Insert mode
            vim.fn.execute('silent! !printf "\\x1b[6 q"', "") -- Steady Bar (I-beam)
        end
      end
    end,
  })
  -- Set initial cursor for Normal mode
  vim.fn.execute('silent! !printf "\\x1b[2 q"', "")


  -- 2. Focus Events (nvim should react if 'gt' sends these)
  --    FocusGained: \\x1b[I
  --    FocusLost:   \\x1b[O
  -- You can create autocommands for these if 'gt' implements sending them.
  -- Example:
  -- vim.api.nvim_create_autocmd("FocusGained", {
  --   pattern = "*",
  --   callback = function()
  --     vim.notify("Neovim gained focus (gt)", vim.log.levels.INFO)
  --     -- Potentially change cursor color or other visual cues
  --   end
  -- })
  -- vim.api.nvim_create_autocmd("FocusLost", {
  --   pattern = "*",
  --   callback = function()
  --     vim.notify("Neovim lost focus (gt)", vim.log.levels.INFO)
  --     -- Potentially change cursor color or other visual cues
  --   end
  -- })


  -- 3. Clipboard (OSC 52)
  -- Instruct Neovim to use OSC 52 for clipboard operations if the 'clipboard' option includes 'osc52'.
  -- This relies on 'gt' correctly handling OSC 52 sequences sent *from* Neovim.
  if vim.fn.has("osc52") == 1 then
    vim.g.clipboard = {
      name = "osc52",
      copy = {
        ["+"] = "\\x1b]52;c;%s\\x07", -- %s is base64 encoded text
        ["*"] = "\\x1b]52;p;%s\\x07", -- %s is base64 encoded text (primary selection)
      },
      paste = {
        -- OSC 52 paste (reading from terminal) is often a security risk and not supported by many terminals.
        -- Neovim itself might not attempt to request paste via OSC 52.
        -- This part is mostly illustrative of how OSC 52 copy is structured.
        ["+"] = "", -- Placeholder, actual paste mechanism might differ
        ["*"] = "", -- Placeholder
      },
      cache_enabled = 0, -- Disable caching if using OSC52
    }
    vim.opt.clipboard = "unnamedplus,osc52"
    vim.notify("gt.lua: Configured clipboard for OSC 52 (requires 'gt' support).", vim.log.levels.INFO)
  else
    vim.notify("gt.lua: OSC 52 clipboard not available in this Neovim build.", vim.log.levels.WARN)
  end


  -- 4. TermOpen Autocommands (General good practices for terminal buffers)
  vim.api.nvim_create_autocmd("TermOpen", {
    pattern = "*",
    group = vim.api.nvim_create_augroup("GtTermOpenSettings", { clear = true }),
    callback = function(args)
      vim.notify("gt.lua: TermOpen event for buffer " .. args.buf, vim.log.levels.INFO)
      -- Disable line numbers in terminal buffers
      vim.opt_local.number = false
      vim.opt_local.relativenumber = false
      -- Ensure 'nomodifiable' is not set if you want to send input
      -- vim.opt_local.modifiable = true -- Usually default for terminal
      -- Optional: Start in insert mode (Terminal-mode)
      -- vim.cmd("startinsert")

      -- Set scrollback if desired (default is 1000 in Neovim IIRC)
      -- vim.opt_local.scrollback = 5000

      -- If 'gt' terminal sends its own title, Neovim might try to set it too.
      -- You can disable Neovim's title setting for terminal buffers if it conflicts.
      -- vim.opt_local.title = false
    end,
  })

  -- Example of how 'gt' might inform Neovim of its name or capabilities
  -- If 'gt' sets an environment variable like GT_VERSION or GT_FEATURES:
  -- local gt_version = vim.env.GT_VERSION
  -- if gt_version then
  --   vim.notify("Running inside gt terminal version: " .. gt_version, vim.log.levels.INFO)
  -- end

  -- Placeholder for more specific 'gt' features as they are developed in 'gt' itself.
  -- For example, if 'gt' supports custom OSC sequences for specific actions.
end

return M 