export const config = { runtime: "edge" };

export default function handler(req) {
  const base = process.env.VITE_API_BASE || "";
  const url = new URL(req.url);
  return Response.redirect(`${base}${url.pathname}${url.search}`, 307);
}
