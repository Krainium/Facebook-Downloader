/** Everything that differs between the two builds lives here. */
export const BRAND = {
  name: "Facebook Downloader",
  short: "Facebook",
  headline: "Save Facebook videos and reels",
  lede:
    "Paste a video, reel or share link and the preview appears on its own. Download the original file in full quality. Nothing is stored on our side.",
  placeholder: "https://www.facebook.com/share/v/…",
  accent: "#4a9bff",
  linkPattern: /(facebook\.com|fb\.watch|fb\.com)\/[^\s]+/i,
  source: "https://github.com/Krainium/Facebook-Downloader",
  steps: [
    { n: "01", t: "Copy the link", d: "Open the video on Facebook and copy its address from the share menu." },
    { n: "02", t: "Paste it here", d: "The preview loads by itself. No button to press, no account, no captcha." },
    { n: "03", t: "Pick a quality", d: "Choose HD or SD and save the file straight to your device." },
  ],
  features: [
    { t: "HD and SD", d: "Both renditions Facebook publishes, so you can trade size against quality." },
    { t: "Videos and reels", d: "Watch pages, reels and the short share links all resolve to the same media." },
    { t: "Nothing is stored", d: "Files stream straight through. No queue, no cache, no copy kept on the server." },
    { t: "No account", d: "No login, no extension, no app. Works the same on desktop and phone." },
  ],
  faq: [
    { q: "Which links work?", a: "Public videos and reels on facebook.com, plus fb.watch and the share links that look like facebook.com/share/v/CODE." },
    { q: "Can it fetch private posts?", a: "No. Only posts visible without logging in can be read, and that limit is deliberate." },
    { q: "What quality do I get?", a: "Whatever Facebook published. Most videos offer an HD and an SD rendition, and both are listed so you can choose." },
    { q: "Do you keep my downloads?", a: "No. The server fetches the file and passes the bytes through as they arrive. Nothing is written to disk." },
  ],
};

/**
 * Palette tokens applied to :root at boot. Facebook blue is dark enough to
 * carry white text, and the canvas leans midnight navy so the blue reads as
 * brand colour rather than as a plain link.
 */
export const THEME: Record<string, string> = {
  "--bg": "#080d16",
  "--ink": "#eaf0fa",
  "--dim": "#93a2ba",
  "--faint": "#64738c",
  "--line": "rgba(120, 165, 255, 0.11)",
  "--line-str": "rgba(120, 165, 255, 0.26)",
  "--panel": "rgba(14, 22, 38, 0.70)",
  "--panel-soft": "rgba(120, 165, 255, 0.05)",
  "--accent": "#1877f2",
  "--accent-2": "#4a9bff",
  "--on-accent": "#ffffff",
  "--glow-1": "rgba(24, 119, 242, 0.26)",
  "--glow-2": "rgba(8, 102, 255, 0.16)",
  "--focus": "rgba(74, 155, 255, 0.55)",
};

/** Wordmark glyph, drawn to sit on the accent chip. */
export const ICON_PATH =
  "M14.2 8.6V7.1c0-.7.3-1.1 1.1-1.1h1.5V3.3h-2.4c-2.2 0-3.4 1.3-3.4 3.5v1.8H8.6v2.7h2.4v9.4h3.2v-9.4h2.4l.4-2.7h-2.8z";
