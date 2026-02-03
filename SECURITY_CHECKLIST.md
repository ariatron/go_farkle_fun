# 🔒 Security Checklist Before Public Upload

Before uploading this repository to GitHub, ensure the following sensitive files are **NOT committed**:

## ⚠️ Files Containing Sensitive Data

### 1. `.env` file
**Status:** ❌ Contains your Grafana Cloud API tokens
**Action:** Already excluded by `.gitignore`
**Safe alternative:** `.env.example` (already sanitized)

**What it contains:**
- Grafana Cloud API token: `glc_eyJvIjoiMTI2NzM5MiI...`
- Prometheus user ID: `1086929`
- Tempo user ID: `1086929`
- Loki user ID: `1086929`

---

### 2. `alloy-config.alloy` file
**Status:** ❌ Contains your Grafana Cloud API tokens
**Action:** Already excluded by `.gitignore`
**Safe alternative:** `alloy-config.alloy.example` (just created with placeholders)

**What it contains:**
- Same Grafana Cloud API token in 3 places (lines 30, 69, 116)
- Prometheus user ID: `1890848`
- Tempo user ID: `1039023`
- Loki user ID: `1044708`

---

## ✅ Pre-Upload Checklist

Before running `git push`:

- [x] `.gitignore` created to exclude sensitive files
- [x] `.env.example` exists with safe placeholder values
- [x] `alloy-config.alloy.example` created with safe placeholder values
- [ ] Verify no credentials in git history: `git log --all --full-history --source -- .env alloy-config.alloy`
- [ ] Double-check before first push: `git status` and ensure `.env` and `alloy-config.alloy` are not staged

---

## 🔄 If You've Already Committed These Files

If you accidentally committed these files to git history:

```bash
# Remove from git history (use with caution!)
git filter-branch --force --index-filter \
  "git rm --cached --ignore-unmatch .env alloy-config.alloy" \
  --prune-empty --tag-name-filter cat -- --all

# Force push to remote (if already pushed)
git push origin --force --all
```

**Better option:** Rotate your Grafana Cloud API tokens at https://grafana.com and create new ones.

---

## 📝 Instructions for Others Using This Repo

Users cloning your repo should:

1. Copy `.env.example` to `.env` and fill in their own credentials
2. Copy `alloy-config.alloy.example` to `alloy-config.alloy` and fill in their credentials
3. Never commit these files to version control

---

## 🆘 If Credentials Are Leaked

1. **Immediately rotate** your Grafana Cloud API token at https://grafana.com
2. Create a new token and update your local files
3. Remove the leaked token from git history (see above)
4. Monitor your Grafana Cloud usage for any suspicious activity
