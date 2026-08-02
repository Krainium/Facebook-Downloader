import { useCallback, useEffect, useRef, useState } from "react";
import SnowBubbles from "./components/SnowBubbles";
import Splash from "./components/Splash";
import { BRAND, ICON_PATH } from "./brand";
import { downloadHref, previewSrc, type Media, type Post, type Variant } from "./types";

type Status = "idle" | "loading" | "ready" | "error";

export default function App() {
  const [link, setLink] = useState("");
  const [status, setStatus] = useState<Status>("idle");
  const [post, setPost] = useState<Post | null>(null);
  const [error, setError] = useState("");
  const [quality, setQuality] = useState(0);
  const [openFaq, setOpenFaq] = useState<number | null>(0);

  // Guards a slow earlier request from overwriting a newer one.
  const requestSeq = useRef(0);
  const lastFetched = useRef("");
  const inputRef = useRef<HTMLInputElement | null>(null);

  const extract = useCallback(async (raw: string) => {
    const value = raw.trim();
    if (!value || value === lastFetched.current) return;
    lastFetched.current = value;

    const seq = ++requestSeq.current;
    setStatus("loading");
    setError("");
    setPost(null);

    try {
      const res = await fetch(`/api/extract?url=${encodeURIComponent(value)}`);
      const data = await res.json();
      if (seq !== requestSeq.current) return; // superseded
      if (!res.ok) {
        setError(data?.error ?? `Request failed (${res.status})`);
        setStatus("error");
        return;
      }
      setPost(data as Post);
      setQuality(0);
      setStatus("ready");
    } catch (e) {
      if (seq !== requestSeq.current) return;
      setError(e instanceof Error ? e.message : "Network error");
      setStatus("error");
    }
  }, []);

  // Auto preview: fire as soon as the field holds a real link, so pasting is the
  // whole interaction. Debounced so typing does not spam the backend.
  useEffect(() => {
    const value = link.trim();
    if (!value) {
      lastFetched.current = "";
      setStatus("idle");
      setPost(null);
      setError("");
      return;
    }
    if (!BRAND.linkPattern.test(value)) return;
    const t = setTimeout(() => void extract(value), 220);
    return () => clearTimeout(t);
  }, [link, extract]);

  const onPaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
    const text = e.clipboardData.getData("text");
    if (text && BRAND.linkPattern.test(text)) {
      e.preventDefault();
      setLink(text.trim());
    }
  };

  const pasteFromClipboard = async () => {
    try {
      const text = await navigator.clipboard.readText();
      if (text) setLink(text.trim());
    } catch {
      /* clipboard blocked, so they can paste manually */
    }
    inputRef.current?.focus();
  };

  const current: Media | undefined = post?.media[0];
  const variants: Variant[] = current?.variants ?? [];
  const chosen: Media | Variant = variants[quality] ?? current!;

  return (
    <>
      <Splash />
      <SnowBubbles />

      <div className="page">
        <header className="nav">
          <a className="brand" href="/">
            <span className="mark" aria-hidden="true">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none">
                <path d={ICON_PATH} stroke="currentColor" strokeWidth="1.5" strokeLinejoin="round" />
              </svg>
            </span>
            <span>{BRAND.name}</span>
          </a>
          <nav className="nav-links">
            <a href="#how">How it works</a>
            <a href="#faq">FAQ</a>
          </nav>
        </header>

        <main>
          <section className="hero">
            <span className="eyebrow">Free · no account</span>
            <h1>{BRAND.headline}</h1>
            <p className="lede">{BRAND.lede}</p>

            <div className={`field ${status === "loading" ? "busy" : ""}`}>
              <span className="field-icon" aria-hidden="true">
                <svg viewBox="0 0 24 24" width="17" height="17" fill="none">
                  <path d="M10 13a5 5 0 007.5.5l2-2a5 5 0 00-7-7l-1 1" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" />
                  <path d="M14 11a5 5 0 00-7.5-.5l-2 2a5 5 0 007 7l1-1" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" />
                </svg>
              </span>
              <input
                ref={inputRef}
                value={link}
                onChange={(e) => setLink(e.target.value)}
                onPaste={onPaste}
                placeholder={BRAND.placeholder}
                spellCheck={false}
                autoComplete="off"
                aria-label={`${BRAND.short} link`}
              />
              {link ? (
                <button className="ghost" onClick={() => setLink("")} aria-label="Clear">✕</button>
              ) : (
                <button className="ghost" onClick={pasteFromClipboard}>Paste</button>
              )}
            </div>

            <p className="assurance">
              {status === "loading" ? "Fetching preview…" : "Public posts only · original quality · nothing saved"}
            </p>
          </section>

          <section className="result" aria-live="polite">
            {status === "error" && <div className="notice error" role="alert">{error}</div>}

            {status === "loading" && !post && (
              <div className="card skeleton">
                <div className="sk-row">
                  <div className="sk-dot" />
                  <div className="sk-lines"><div className="sk-line" /><div className="sk-line short" /></div>
                </div>
                <div className="sk-frame" />
              </div>
            )}

            {post && current && (
              <article className="card">
                {(post.author || post.handle) && (
                  <div className="who">
                    {post.avatar && <img className="avatar" src={post.avatar} alt="" loading="lazy" />}
                    <div className="who-text">
                      <strong>{post.author || post.handle}</strong>
                      {post.handle && <span>{post.handle.startsWith("@") ? post.handle : `@${post.handle}`}</span>}
                    </div>
                  </div>
                )}

                {post.title && <p className="tweet-text">{post.title}</p>}

                <div className="stage">
                  <video
                    className="media"
                    src={previewSrc(chosen)}
                    poster={current.thumb ? previewSrc({ ...current, url: current.thumb } as Media) : undefined}
                    controls
                    playsInline
                    preload="metadata"
                  />
                </div>

                {variants.length > 1 && (
                  <div className="quality">
                    <span className="quality-label">Quality</span>
                    <div className="quality-opts" role="radiogroup" aria-label="Video quality">
                      {variants.map((v, i) => (
                        <button
                          key={v.url}
                          role="radio"
                          aria-checked={i === quality}
                          className={`q ${i === quality ? "on" : ""}`}
                          onClick={() => setQuality(i)}
                        >
                          <span className="q-res">{v.label}</span>
                        </button>
                      ))}
                    </div>
                  </div>
                )}

                <div className="actions">
                  <a className="cta" href={downloadHref(chosen)}>Download video</a>
                </div>
              </article>
            )}
          </section>

          <section id="how" className="band">
            <h2>How it works</h2>
            <div className="steps">
              {BRAND.steps.map((s) => (
                <div className="step" key={s.n}>
                  <span className="step-n">{s.n}</span>
                  <h3>{s.t}</h3>
                  <p>{s.d}</p>
                </div>
              ))}
            </div>
          </section>

          <section className="band">
            <h2>What you get</h2>
            <div className="grid">
              {BRAND.features.map((f) => (
                <div className="tile" key={f.t}><h3>{f.t}</h3><p>{f.d}</p></div>
              ))}
            </div>
          </section>

          <section id="faq" className="band">
            <h2>Questions</h2>
            <div className="faq">
              {BRAND.faq.map((item, i) => (
                <div className={`qa ${openFaq === i ? "open" : ""}`} key={item.q}>
                  <button onClick={() => setOpenFaq(openFaq === i ? null : i)} aria-expanded={openFaq === i}>
                    <span>{item.q}</span>
                    <span className="chev" aria-hidden="true">⌄</span>
                  </button>
                  <div className="qa-body"><p>{item.a}</p></div>
                </div>
              ))}
            </div>
          </section>
        </main>

        <footer className="foot">
          <p>Not affiliated with {BRAND.short}.</p>
        </footer>
      </div>
    </>
  );
}
