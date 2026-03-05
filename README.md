# go-chat

Real-time chat with Go, WebSockets, Google OAuth, and MongoDB Atlas.

**Live:** https://go-chat-5ggtkeu3aq-ts.a.run.app

---

## Local development

**1. Create `.env`**

```env
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
APP_BASE_URL=http://localhost:8080
OAUTH_REDIRECT_URL=http://localhost:8080/auth/google/callback
MONGO_URI=mongodb+srv://<user>:<password>@<cluster>.mongodb.net/?appName=go-chat
```

**2. Add** `http://localhost:8080/auth/google/callback` **to Google OAuth authorised redirect URIs.**

**3. Allow your IP in MongoDB Atlas → Network Access.**

**4. Run**

```bash
make run
```

---

## Commands

| Command       | Description             |
| ------------- | ----------------------- |
| `make run`    | Run locally             |
| `make build`  | Compile to `bin/server` |
| `make test`   | Run tests               |
| `make deploy` | Deploy to Cloud Run     |
| `make clean`  | Remove `bin/`           |

---

## Deploy to Google Cloud Run

```bash
gcloud auth login
gcloud config set project YOUR_PROJECT_ID
make deploy
```

The script handles: build → APIs → IAM → secrets → Cloud Run deploy.

**Atlas:** set Network Access to `0.0.0.0/0` (Cloud Run has dynamic IPs).

---

## Environment variables

| Variable               | Default                             | Description                                  |
| ---------------------- | ----------------------------------- | -------------------------------------------- |
| `GOOGLE_CLIENT_ID`     | required                            | OAuth client ID                              |
| `GOOGLE_CLIENT_SECRET` | required                            | OAuth client secret                          |
| `MONGO_URI`            | `mongodb://localhost:27017`         | MongoDB URI                                  |
| `OAUTH_REDIRECT_URL`   | `APP_BASE_URL/auth/google/callback` | Full OAuth redirect URL                      |
| `APP_BASE_URL`         | `http://localhost:<port>`           | Base URL for redirect URI                    |
| `SESSION_SECRET`       | random                              | Cookie signing key — set to persist sessions |
| `MONGO_DB`             | `gochat`                            | Database name                                |
| `HISTORY_LIMIT`        | `50`                                | Messages loaded on connect                   |
| `PORT`                 | `8080`                              | Injected by Cloud Run                        |
