/*jslint browser,long,fart,indent2*/
/*global alert,console,window,WebSocket,JSON*/

(function () {
  "use strict";

  const mermaidQuery = "code.language-mermaid";
  const copyIcon = `<svg class="copy-icon" aria-hidden="true" fill="none" height="18" shape-rendering="geometricPrecision" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" viewBox="0 0 24 24" width="18" style="color:"currentColor";"><path d="M8 17.929H6c-1.105 0-2-.912-2-2.036V5.036C4 3.91 4.895 3 6 3h8c1.105 0 2 .911 2 2.036v1.866m-6 .17h8c1.105 0 2 .91 2 2.035v10.857C20 21.09 19.105 22 18 22h-8c-1.105 0-2-.911-2-2.036V9.107c0-1.124.895-2.036 2-2.036z"></path></svg>`;
  const tickIcon = `<svg class="tick-icon" aria-hidden="true" fill="none" height="18" shape-rendering="geometricPrecision" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" viewBox="0 0 24 24" width="18" style="color: "currentColor";"><path d="M5 13l4 4L19 7"></path></svg>`;

  function loadMermaid(isLight) {
    const theme = (
      isLight
      ? "default"
      : "dark"
    );
    window.mermaid.initialize({startOnLoad: false, theme});
    window.mermaid.run({querySelector: mermaidQuery});
  }

  function saveOriginalData() {
    return new Promise((resolve, reject) => {
      try {
        const els = document.querySelectorAll(mermaidQuery);
        els.forEach((element) => {
          element.setAttribute("data-original-code", element.innerHTML);
        });
        resolve();
      } catch (error) {
        reject(error);
      }
    });
  }

  function resetProcessed() {
    return new Promise((resolve, reject) => {
      try {
        const els = document.querySelectorAll(mermaidQuery);
        els.forEach((element) => {
          if (element.getAttribute("data-original-code") !== null) {
            element.removeAttribute("data-processed");
            element.innerHTML = element.getAttribute("data-original-code");
          }
        });
        resolve();
      } catch (error) {
        reject(error);
      }
    });
  }

  function initMermaid() {
    // Workaround issue with MermaidJS that doesn't allow changing theme on
    // re-initialization
    // https://github.com/mermaid-js/mermaid/issues/1945#issuecomment-1661264708
    saveOriginalData().catch(console.error);

    if (window.Param.mode === "dark") {
      loadMermaid(true);
    } else if (window.Param.mode === "light") {
      loadMermaid(false);
    } else {
      const prefersLightQuery = window.matchMedia("(prefers-color-scheme: light)");
      loadMermaid(prefersLightQuery.matches);
      // Change CSS when the theme changes
      prefersLightQuery.addEventListener("change", (e) => {
        resetProcessed().then(loadMermaid(e.matches)).catch(console.error);
      });
    }
  }

  async function loadMarkdown() {
    const response = await fetch(`/__/md?path=${window.location.pathname.slice(1)}`);
    const result = await response.json();

    const markdownBody = document.getElementById("markdown-body");
    markdownBody.innerHTML = result.html;

    const markdownTitle = document.getElementById("markdown-title");
    markdownTitle.innerHTML = result.title;

    initMermaid();
    await typesetMathJax();
    addCopyButtons();
    buildHeadingsList();
  }

  async function typesetMathJax() {
    if (window.MathJax) {
      try {
        if (window.MathJax.startup && window.MathJax.startup.promise) {
          await window.MathJax.startup.promise;
        }
        if (window.MathJax.typesetPromise) {
          await window.MathJax.typesetPromise();
        } else if (window.MathJax.typeset) {
          window.MathJax.typeset();
        }
      } catch (error) {
        console.error(error);
      }
    }
  }

  function addCopyButtons() {
    document.querySelectorAll(".markdown-body pre").forEach((pre) => {
      const code = pre.querySelector("code");
      const button = document.createElement("button");
      button.classList.add("copy-button");
      button.setAttribute("aria-label", "Copy code to clipboard");
      button.innerHTML = copyIcon;
      pre.style.position = "relative";
      pre.appendChild(button);

      button.addEventListener("click", () => {
        navigator.clipboard.writeText(code.textContent).then(() => {
          button.innerHTML = tickIcon;
          setTimeout(() => {
            button.innerHTML = copyIcon;
          }, 1000);
        }).catch(() => {
          alert("Failed to copy");
        });
      });
    });
  }

  function slugify(text) {
    return text
      .toLowerCase()
      .trim()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/(^-|-$)+/g, "");
  }

  function buildHeadingsList() {
    const details = document.getElementById("heading-list");
    const list = document.getElementById("headings-tree");
    const markdownBody = document.getElementById("markdown-body");

    if (!details || !list || !markdownBody) {
      if (details) {
        details.classList.add("hidden");
      }
      return;
    }

    const headings = Array.from(
      markdownBody.querySelectorAll("h1, h2, h3, h4, h5, h6")
    );
    list.innerHTML = "";

    if (headings.length === 0) {
      details.classList.add("hidden");
      return;
    }

    details.classList.remove("hidden");

    const usedIds = new Map();
    headings.forEach((heading) => {
      const level = Number(heading.tagName.slice(1));
      const text = heading.textContent.trim();
      if (!text) {
        return;
      }

      let id = heading.id;
      if (!id) {
        const base = slugify(text) || "heading";
        let candidate = base;
        let index = 1;
        while (usedIds.has(candidate) || document.getElementById(candidate)) {
          index += 1;
          candidate = `${base}-${index}`;
        }
        id = candidate;
        heading.id = id;
      }

      usedIds.set(id, true);

      const link = document.createElement("a");
      link.href = `#${id}`;
      link.className = `heading-item heading-level-${level}`;
      link.textContent = text;
      link.addEventListener("click", () => {
        details.open = false;
      });
      list.appendChild(link);
    });
  }

  (async function () {
    // Only load markdown initially if not in directory index mode
    if (!window.Param.isDirectoryIndex) {
      await loadMarkdown();
    }

    if (window.Param.reload) {
      const conn = new WebSocket(`ws://${window.Param.host}/ws`);
      conn.onopen = () => conn.send("Ping");
      conn.onerror = (e) => console.log(`Connection error: ${e}`);
      conn.onclose = (e) => console.log(`Connection closed: ${e}`);
      conn.onmessage = (e) => {
        if (e.data === "reload") {
          console.log("Reload page!");
          // For directory index view, do a full page reload
          // For markdown view, reload just the markdown content
          if (window.Param.isDirectoryIndex) {
            window.location.reload();
          } else {
            loadMarkdown();
          }
        }
      };
    }
  }());
}());

// Popover functionality
document.addEventListener("DOMContentLoaded", function () {
  const popovers = [
    document.getElementById("file-browser"),
    document.getElementById("heading-list"),
  ].filter(Boolean);

  if (popovers.length === 0) {
    return;
  }

  document.addEventListener("click", function (e) {
    popovers.forEach((details) => {
      if (!details.open) {
        return;
      }

      if (details.contains(e.target)) {
        return;
      }

      details.open = false;
    });
  });
});
