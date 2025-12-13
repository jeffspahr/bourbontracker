# Cloudflare Deployment Guide

Deploy the Bourbon Tracker to Cloudflare Pages with Google OAuth authentication using Cloudflare Access.

### Prerequisites

- Cloudflare account (Free tier works)
- Domain managed by Cloudflare
- Google OAuth credentials

### Setup Steps

#### 1. Deploy to Cloudflare Pages

```bash
# Install Wrangler CLI
npm install -g wrangler

# Login to Cloudflare
wrangler login

# Deploy
wrangler pages deploy . --project-name=bourbon-tracker
```

#### 2. Configure Google OAuth

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project (or use existing)
3. Enable Google+ API
4. Create OAuth 2.0 credentials:
   - Application type: Web application
   - Authorized redirect URIs: `https://<your-team>.cloudflareaccess.com/cdn-cgi/access/callback`
5. Save Client ID and Client Secret

#### 3. Set Up Cloudflare Access

1. Go to Cloudflare Dashboard → Zero Trust → Access → Applications
2. Click "Add an application" → "Self-hosted"
3. Configure:
   - **Application name**: Bourbon Tracker
   - **Subdomain**: bourbon-tracker
   - **Domain**: your-domain.com
   - **Path**: (leave empty to protect entire site)

4. Add Identity Provider:
   - Go to Settings → Authentication
   - Click "Add new" → Google
   - Enter Client ID and Client Secret from Google
   - Save

5. Create Access Policy:
   - **Policy name**: Allowed Users
   - **Action**: Allow
   - **Include**: Emails
   - Add allowed email addresses:
     ```
     user1@gmail.com
     user2@gmail.com
     admin@yourdomain.com
     ```

6. Save and deploy

#### 4. Access Your Site

Visit `https://bourbon-tracker.your-domain.com`

Users will be prompted to sign in with Google, and only allowlisted emails can access.

### Cost

- **Free tier**: Up to 50 users
- **Zero Trust Free**: Includes Cloudflare Access
- No credit card required

---

## Updating Inventory

The tracker generates `inventory.json` locally. To update the hosted version:

### Option A: GitHub Actions (Automated)

Add a workflow to run the tracker and deploy to Cloudflare Pages on schedule.

### Option B: Manual Upload

```bash
# Run tracker locally
./tracker -va -wake

# Deploy updated inventory
wrangler pages deploy . --project-name=bourbon-tracker
```

### Option C: Cloudflare Workers Cron

Run the tracker as a scheduled Worker (requires Docker or Go build in Workers).

---

## Security Notes

- Cloudflare Access is SOC 2 Type II certified
- OAuth tokens are managed by Cloudflare
- Email allowlist can be updated anytime via dashboard
- Supports 2FA if enabled on Google accounts
- Rate limiting and DDoS protection included

## Troubleshooting

**"Access Denied" after signing in:**
- Check email is in allowlist
- Verify email matches Google account exactly
- Check Access policy is set to "Allow"

**OAuth redirect error:**
- Verify redirect URI in Google Console matches Cloudflare Access
- Format: `https://<your-team>.cloudflareaccess.com/cdn-cgi/access/callback`

**Page not loading:**
- Check Cloudflare Pages build logs
- Verify `inventory.json` and `map.html` are in deployment
- Check browser console for errors
