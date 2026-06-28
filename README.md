# QazFreelance

A Telegram bot for submitting freelance orders and self-promotion posts (resumes/advertisements). 

## Prerequisites

- Go 1.22 or later

## Setup

Clone the repository and download dependencies:

```sh
git clone https://github.com/AnmiTaliDev/qazfreelance.git
cd qazfreelance
go mod download
```

## Environment Variables

| Variable        | Required | Description                                                     |
|-----------------|----------|-----------------------------------------------------------------|
| `BOT_TOKEN`     | Yes      | Telegram bot token from @BotFather                              |
| `MODERATOR_IDS` | No       | Comma-separated list of moderator Telegram user IDs             |
| `CHANNEL_ID`    | Yes      | Telegram channel ID where approved posts are published          |

Example:

```sh
export BOT_TOKEN=123456789:ABCDefghIJKlmnoPQRstuvWXYZ
export MODERATOR_IDS=111111111,222222222
export CHANNEL_ID=-1001234567890
```

## Running the Bot

```sh
go run ./cmd/bot
```

The bot creates a `qazfreelance.db` SQLite database file in the current directory on first run.

## Usage

### For Users

1. Start the bot with `/start`.
2. Choose your language (Kazakh, Russian, or English).
3. From the main menu, choose either:
   - **Post an order** — if you need a freelancer for a task.
   - **Advertise yourself** — if you offer services.
4. Fill in the title, description, and contact information step by step.
5. Your submission is sent for moderation. You will be notified when it is approved or rejected.

### For Moderators

Moderators see an additional **Moderator menu** button in the main menu, and can use `/mode` to toggle between modes.

**Stream mode** (default): new submissions appear immediately as messages with Approve/Reject buttons.

**List mode**: submissions queue silently. A notification appears with an "Open list" button to browse and act on pending items one by one.

#### Decision Logic

- Any moderator approves, none reject -> submission is published to the channel.
- Any moderator rejects, none approve -> submission is rejected and the user is notified.
- One approves and another rejects -> submission moves to **Conflicts**. Any moderator can open the Conflicts list and make a final decision.

## License

GNU Affero General Public License v3.0 only (AGPL-3.0-only).
