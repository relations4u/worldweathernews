/**
 * GitHub OAuth proxy for Sveltia CMS.
 *
 * Sveltia CMS (and its predecessor Decap CMS) speak a postMessage-based OAuth
 * handshake. The browser opens a popup at <auth-base-url>/auth, which redirects
 * to GitHub. After GitHub authorises, it redirects back to <auth-base-url>/callback
 * with a code. This worker exchanges that code for a personal access token and
 * delivers it to the opener via postMessage. The token then sits in Sveltia's
 * IndexedDB and is used directly against the GitHub API.
 *
 * Adapted from https://github.com/sveltia/sveltia-cms-auth (MIT). The flow is
 * intentionally minimal: only `provider=github` is supported, allowed origins
 * are validated explicitly, CSRF state is bound via a Secure HttpOnly cookie.
 */

export interface Env {
  GITHUB_CLIENT_ID: string;
  GITHUB_CLIENT_SECRET: string;
  /** Comma-separated list of bare hostnames allowed as Sveltia origins. */
  ALLOWED_DOMAINS: string;
}

const renderResponseHTML = (
  status: "success" | "error",
  payload: unknown,
): string => {
  const data = JSON.stringify(payload);
  return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Authorising…</title>
</head>
<body>
<script>
(function() {
  var data = ${data};
  function receive(event) {
    if (event.data === 'authorizing:github') {
      window.removeEventListener('message', receive, false);
      window.opener.postMessage(
        'authorization:github:${status}:' + JSON.stringify(data),
        event.origin
      );
    }
  }
  window.addEventListener('message', receive, false);
  window.opener.postMessage('authorizing:github', '*');
})();
</script>
</body>
</html>`;
};

const isAllowedOrigin = (rawOrigin: string, allowed: string[]): boolean => {
  if (!rawOrigin) return false;
  let host: string;
  try {
    host = new URL(
      rawOrigin.startsWith("http") ? rawOrigin : `https://${rawOrigin}`,
    ).hostname;
  } catch {
    return false;
  }
  return allowed.some((d) => host === d || host.endsWith("." + d));
};

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    const allowed = env.ALLOWED_DOMAINS.split(",")
      .map((s) => s.trim())
      .filter(Boolean);

    if (url.pathname === "/auth") {
      const provider = url.searchParams.get("provider");
      if (provider !== "github") {
        return new Response("Unsupported provider", { status: 400 });
      }
      const siteId = url.searchParams.get("site_id") || "";
      if (!isAllowedOrigin(siteId, allowed)) {
        return new Response("Origin not allowed", { status: 403 });
      }
      const scope = url.searchParams.get("scope") || "repo,user";
      const state = crypto.randomUUID();
      const ghUrl = new URL("https://github.com/login/oauth/authorize");
      ghUrl.searchParams.set("client_id", env.GITHUB_CLIENT_ID);
      ghUrl.searchParams.set("scope", scope);
      ghUrl.searchParams.set("state", state);
      ghUrl.searchParams.set("redirect_uri", `${url.origin}/callback`);
      return new Response(null, {
        status: 302,
        headers: {
          Location: ghUrl.toString(),
          "Set-Cookie": `csrf-token=${state}; Path=/; SameSite=Lax; Secure; HttpOnly; Max-Age=600`,
        },
      });
    }

    if (url.pathname === "/callback") {
      const code = url.searchParams.get("code");
      const state = url.searchParams.get("state");
      const cookie = request.headers.get("cookie") || "";
      const csrfMatch = cookie.match(/csrf-token=([^;]+)/);
      if (!code || !state || !csrfMatch || csrfMatch[1] !== state) {
        return new Response(
          renderResponseHTML("error", { error: "invalid state" }),
          {
            status: 400,
            headers: { "Content-Type": "text/html; charset=utf-8" },
          },
        );
      }
      const tokenResp = await fetch(
        "https://github.com/login/oauth/access_token",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Accept: "application/json",
          },
          body: JSON.stringify({
            client_id: env.GITHUB_CLIENT_ID,
            client_secret: env.GITHUB_CLIENT_SECRET,
            code,
          }),
        },
      );
      const tokenData = (await tokenResp.json()) as {
        access_token?: string;
        error?: string;
        error_description?: string;
      };
      if (!tokenResp.ok || !tokenData.access_token) {
        return new Response(
          renderResponseHTML("error", {
            error: tokenData.error || "token exchange failed",
            description: tokenData.error_description,
          }),
          {
            status: 401,
            headers: { "Content-Type": "text/html; charset=utf-8" },
          },
        );
      }
      return new Response(
        renderResponseHTML("success", {
          provider: "github",
          token: tokenData.access_token,
        }),
        {
          status: 200,
          headers: {
            "Content-Type": "text/html; charset=utf-8",
            "Set-Cookie": "csrf-token=; Path=/; Max-Age=0",
          },
        },
      );
    }

    if (url.pathname === "/") {
      return new Response("wwn-cms-auth: GitHub OAuth proxy for Sveltia CMS", {
        status: 200,
        headers: { "Content-Type": "text/plain; charset=utf-8" },
      });
    }

    return new Response("Not found", { status: 404 });
  },
};
