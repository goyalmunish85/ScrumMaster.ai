# Integration Setup Guide

To connect the **AI Delivery OS** to your external platforms securely, you will need to generate API credentials for each tool and add them to your `backend/.env` file. 

The system uses these credentials in **Read-Only** mode to pull context. It will never write data back or send messages on your behalf.

---

## The Configurations File
All *targets* (like the specific Google Sheet IDs, Slack Channel IDs, or Jira Project Keys) are defined in the `backend/integrations.json` file.

```json
{
  "slack_channels": ["C012345", "C098765"],
  "jira_projects": ["ENG", "OPS"],
  "gitlab_projects": ["12345"],
  "google_sheets": ["1m1LxPmbxt-ZRxeib3bnyAcqaCFzgKJkbqpSJQc08PFc"]
}
```

The actual secret *credentials* (API Tokens) are defined in the `backend/.env` file.

---

## 1. Google Sheets
**Complexity:** Easy

Google Sheets is the easiest to set up. We use the public CSV export URL to read the data without needing complex OAuth.
1. Open your Google Sheet.
2. Click **Share** (top right) and ensure General Access is set to **"Anyone with the link can view"**.
3. Look at the URL in your browser. It looks like this:
   `https://docs.google.com/spreadsheets/d/1m1LxPmbxt-ZRxeib3bnyAcqaCFzgKJkbqpSJQc08PFc/edit`
4. Copy the long ID string (`1m1LxPmbxt-ZRxeib3bnyAcqaCFzgKJkbqpSJQc08PFc`).
5. Open `backend/integrations.json` and add this ID to the `"google_sheets"` array.

---

## 2. Slack
**Complexity:** Medium

We need a Bot Token to read a specific channel's history.
1. Go to [https://api.slack.com/apps](https://api.slack.com/apps) and click **Create New App** > **From Scratch**.
2. Give it a name (e.g., "AI Delivery OS") and pick your workspace.
3. In the left sidebar, click **OAuth & Permissions**.
4. Scroll down to **Scopes** -> **Bot Token Scopes** and add:
   - `channels:history`
   - `channels:read`
5. Scroll up and click **Install to Workspace**.
6. Copy the **Bot User OAuth Token** (starts with `xoxb-`) and paste it into `backend/.env` as `SLACK_BOT_TOKEN`.
7. **Find your Channel ID**: Open Slack, right-click the channel you want the AI to read, select **View channel details**, scroll to the bottom, and copy the **Channel ID** (e.g., `C0123456789`). Add this to the `"slack_channels"` array in `integrations.json`.
8. *Important*: Go to that channel in Slack and type `/invite @AI Delivery OS` so the bot has permission to read it.

---

## 3. Jira
**Complexity:** Medium

We need an API token to pull your Jira tickets.
1. Go to your Atlassian Account Security settings: [https://id.atlassian.com/manage-profile/security/api-tokens](https://id.atlassian.com/manage-profile/security/api-tokens)
2. Click **Create API token**, name it "AI OS", and copy the token.
3. Open `backend/.env` and fill out:
   - `JIRA_DOMAIN=yourcompany.atlassian.net`
   - `JIRA_EMAIL=your.email@company.com` (The email you use to log into Jira)
   - `JIRA_API_TOKEN=` (Paste the token you just created)
4. Add your Jira Project Keys (e.g., `ENG`, `PROD`) to the `"jira_projects"` array in `integrations.json`.

---

## 4. GitLab
**Complexity:** Easy

We need a Personal Access Token to read Merge Requests.
1. In GitLab, click your avatar (top right) -> **Edit profile** -> **Access Tokens** (left sidebar).
2. Click **Add new token**.
3. Name it "AI OS", uncheck expiration, and check the `read_api` scope.
4. Copy the token (starts with `glpat-`) and put it in `backend/.env` as `GITLAB_ACCESS_TOKEN`.
5. Open `backend/.env` and fill out `GITLAB_DOMAIN=gitlab.com` (or your self-hosted domain).
6. **Find your Project ID**: Go to your GitLab project's homepage. Below the project name, you will see a **Project ID** number. Add this to the `"gitlab_projects"` array in `integrations.json`.

---

Once you have filled out your `.env` credentials and listed your targets in `integrations.json`, the AI OS will immediately be able to pull data from all these sources concurrently when you click **"Sync Integrations"**!
