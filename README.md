# ssm-sync

Self-hosted sync server for [ssm](https://github.com/Dilgo-dev/ssm) (SSH connection manager).

Zero-knowledge: the server stores encrypted blobs only, it never sees your plaintext data.

## Quick start (Docker)

```bash
docker compose up -d
```

Then in ssm, set the server to `http://your-server:8080` during `ssm register` or `ssm login`.

## Quick start (binary)

Download from [Releases](https://github.com/Dilgo-dev/ssm-sync/releases), then:

```bash
JWT_SECRET=your-secret-here ./ssm-sync
```

## Configuration

All configuration is done via environment variables.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JWT_SECRET` | Yes | — | Secret key for signing auth tokens |
| `PORT` | No | `8080` | Server port |
| `DATA_DIR` | No | `./data` | Directory for SQLite database |
| `SMTP_HOST` | No | — | SMTP server (if empty, email verification is skipped) |
| `SMTP_PORT` | No | `587` | SMTP port |
| `SMTP_USER` | No | — | SMTP username / from address |
| `SMTP_PASS` | No | — | SMTP password |
| `API_URL` | No | `http://localhost:PORT` | Public URL (used in verification emails) |

## Email verification

By default, email verification is **disabled**. Accounts are immediately active.

To enable it, set `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, and `SMTP_PASS`. Users will then need to verify their email before they can sync.

## CLI setup

```
ssm register
```

Change the **Server** field to your self-hosted URL (e.g. `http://192.168.1.50:8080`).

## API

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/auth/register` | — | Create account |
| POST | `/auth/login` | — | Login |
| GET | `/auth/status` | Bearer | Check email verification status |
| GET | `/auth/verify?token=X` | — | Verify email (link from email) |
| POST | `/auth/resend-verification` | — | Resend verification email |
| GET | `/sync` | Bearer | Download encrypted vault |
| PUT | `/sync` | Bearer | Upload encrypted vault |
| GET | `/health` | — | Health check |

## Build from source

```bash
go build ./cmd/ssm-sync
```

## License

MIT
