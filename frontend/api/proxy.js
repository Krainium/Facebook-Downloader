export const config = { runtime: "edge" };

export default function handler(req) {
  const base = process.env.VITE_API_BASE || "";
  const url = new URL(req.url);
  const p = url.searchParams.get("p") || "";
  url.searchParams.delete("p");
  const qs = url.searchParams.toString();
  const search = qs ? `?${qs}` : "";
  return Response.redirect(`${base}/api/${p}${search}`, 307);
}
