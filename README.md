<div align="center">

# 📘 Facebook Downloader

### Save Facebook videos and reels in HD or SD. No account, no app, nothing stored.

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-20232A?style=for-the-badge&logo=react&logoColor=61DAFB)
![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=for-the-badge&logo=typescript&logoColor=white)
![Vite](https://img.shields.io/badge/Vite-646CFF?style=for-the-badge&logo=vite&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-1877F2?style=for-the-badge&logoColor=white)

<img src="docs/preview.png" alt="Facebook Downloader preview" width="820" />

</div>

---

## ✨ Features

| | |
|:--|:--|
| ⚡ | **Auto preview** &nbsp; Paste a link and the video appears on its own, no button to press |
| 🎚️ | **HD and SD** &nbsp; Both renditions Facebook publishes, pick size against quality |
| 🎬 | **Videos and reels** &nbsp; Watch pages, reels and short share links all resolve to the same media |
| 🔒 | **Nothing stored** &nbsp; Files stream straight through, no queue, no cache, no copy on the server |
| 👤 | **No account** &nbsp; No login, no extension, no app, the same on desktop and phone |
| 🎯 | **Right video every time** &nbsp; The requested reel is pinned by its id, never a recommendation |

## 🧭 How it works

```
🔗 Copy the link  ➜  📥 Paste it here  ➜  🎚️ Pick a quality
```

1. 🔗 **Copy the link** &nbsp; Open the video on Facebook and copy its address from the share menu
2. 📥 **Paste it here** &nbsp; The preview loads by itself, no captcha and no sign in
3. 🎚️ **Pick a quality** &nbsp; Choose HD or SD and save the file straight to your device

.

## 🗂️ Project layout

```
facebook-downloader/
├── 🟢 backend/     Go server, extraction, media proxy, static host
│   └── main.go
└── 🔵 frontend/    React, TypeScript, Vite
    └── src/
        ├── App.tsx
        ├── brand.ts       theme and copy
        ├── components/     splash, snow bubbles
        └── styles.css
```

## 🔌 API

| Method | Endpoint | Purpose |
|:------:|:--|:--|
| `GET` | `/api/health` | Liveness and proxy status |
| `GET` | `/api/extract?url=` | Resolve a link to the video, HD and SD |
| `GET` | `/api/media?url=` | Inline playback stream for the preview |
| `GET` | `/api/download?url=&filename=` | Download stream with a real filename |

## 🚀 Run it

**Backend**

```bash
cd backend
go run .            # listens on :4447, serves ./public
```

**Frontend**

```bash
cd frontend
npm install
npm run dev         # Vite dev server
npm run build       # production build into dist/
```

### 🔧 Environment

| Variable | Default | Meaning |
|:--|:--|:--|
| `PORT` | `4447` | Port the server listens on |
| `FDL_DIST` | `./public` | Folder of built frontend assets |
| `FDL_PROXY` | none | Optional upstream proxy for outbound fetches |



<div align="center">

Made for saving the clips you want to keep. 📘💙

📄 MIT License

</div>
