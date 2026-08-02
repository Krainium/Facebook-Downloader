export interface Variant {
  url: string;
  label: string;
  filename: string;
}

export interface Media {
  type: "video" | "photo";
  url: string;
  thumb: string;
  ext: string;
  filename: string;
  /** Present only when the platform offers more than one rendition. */
  variants?: Variant[];
}

export interface Post {
  id: string;
  title: string;
  author: string;
  handle: string;
  avatar: string;
  pageUrl: string;
  media: Media[];
}

export const downloadHref = (m: Media | Variant): string =>
  `/api/download?url=${encodeURIComponent(m.url)}&filename=${encodeURIComponent(m.filename)}`;

/**
 * Inline playback url. The CDN rejects browser requests that arrive without the
 * platform's own referer, so previews go through the server rather than
 * straight to the CDN.
 */
export const previewSrc = (m: Media | Variant): string =>
  `/api/media?url=${encodeURIComponent(m.url)}`;
