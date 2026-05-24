/*jslint browser,long,fart,indent2*/
/*global alert,console,window,WebSocket,JSON*/

(function () {
  "use strict";

  const mermaidQuery = "code.language-mermaid";
  const mermaidMinScale = 0.2;
  const mermaidMaxScale = 5;
  const mermaidZoomInFactor = 1.1;
  const mermaidZoomOutFactor = 0.9;
  const mermaidOverlayPadding = 64;
  const geoJSONQuery = "code.language-geojson";
  const topoJSONQuery = "code.language-topojson";
  const mapQuery = `${geoJSONQuery}, ${topoJSONQuery}`;
  const copyIcon = `<svg class="copy-icon" aria-hidden="true" fill="none" height="18" shape-rendering="geometricPrecision" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" viewBox="0 0 24 24" width="18" style="color:"currentColor";"><path d="M8 17.929H6c-1.105 0-2-.912-2-2.036V5.036C4 3.91 4.895 3 6 3h8c1.105 0 2 .911 2 2.036v1.866m-6 .17h8c1.105 0 2 .91 2 2.035v10.857C20 21.09 19.105 22 18 22h-8c-1.105 0-2-.911-2-2.036V9.107c0-1.124.895-2.036 2-2.036z"></path></svg>`;
  const tickIcon = `<svg class="tick-icon" aria-hidden="true" fill="none" height="18" shape-rendering="geometricPrecision" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" viewBox="0 0 24 24" width="18" style="color: "currentColor";"><path d="M5 13l4 4L19 7"></path></svg>`;
  const expandIcon = `<svg class="expand-icon" aria-hidden="true" viewBox="0 0 1792 1792" width="14" height="14" fill="currentColor"><path d="M883 1056q0 13-10 23l-332 332 144 144q19 19 19 45t-19 45-45 19h-448q-26 0-45-19t-19-45v-448q0-26 19-45t45-19 45 19l144 144 332-332q10-10 23-10t23 10l114 114q10 10 10 23zm781-864v448q0 26-19 45t-45 19-45-19l-144-144-332 332q-10 10-23 10t-23-10l-114-114q-10-10-10-23t10-23l332-332-144-144q-19-19-19-45t19-45 45-19h448q26 0 45 19t19 45z"></path></svg>`;
  let diagramMediaQuery;
  let overlayCleanup;
  let overlayEl;

  async function loadMermaid(isLight) {
    const theme = (
      isLight
      ? "default"
      : "dark"
    );
    window.mermaid.initialize({startOnLoad: false, theme});
    await window.mermaid.run({querySelector: mermaidQuery});
    setupMermaidPanZoom();
  }

  function attachPanZoom(element, svg) {
    const state = {dragging: false, lastX: 0, lastY: 0, scale: 1, tx: 0, ty: 0};
    const applyTransform = () => {
      svg.style.transform = `translate(${state.tx}px, ${state.ty}px) scale(${state.scale})`;
    };
    const onWheel = (e) => {
      const rect = element.getBoundingClientRect();
      const factor = (
        e.deltaY < 0
        ? mermaidZoomInFactor
        : mermaidZoomOutFactor
      );
      const mx = e.clientX - rect.left;
      const my = e.clientY - rect.top;
      const newScale = Math.max(mermaidMinScale, Math.min(mermaidMaxScale, state.scale * factor));
      e.preventDefault();
      if (newScale === state.scale) {
        return;
      }
      state.tx = mx - (mx - state.tx) * (newScale / state.scale);
      state.ty = my - (my - state.ty) * (newScale / state.scale);
      state.scale = newScale;
      applyTransform();
    };
    const onPointerDown = (e) => {
      if (e.button !== 0) {
        return;
      }
      state.dragging = true;
      state.lastX = e.clientX;
      state.lastY = e.clientY;
      element.style.cursor = "grabbing";
      element.setPointerCapture(e.pointerId);
      e.preventDefault();
    };
    const onPointerMove = (e) => {
      if (!state.dragging) {
        return;
      }
      state.tx += e.clientX - state.lastX;
      state.ty += e.clientY - state.lastY;
      state.lastX = e.clientX;
      state.lastY = e.clientY;
      applyTransform();
    };
    const onPointerUp = () => {
      if (!state.dragging) {
        return;
      }
      state.dragging = false;
      element.style.cursor = "grab";
    };
    const detach = () => {
      element.removeEventListener("wheel", onWheel);
      element.removeEventListener("pointerdown", onPointerDown);
      element.removeEventListener("pointermove", onPointerMove);
      element.removeEventListener("pointerup", onPointerUp);
      element.removeEventListener("pointercancel", onPointerUp);
    };
    element.addEventListener("wheel", onWheel, {passive: false});
    element.addEventListener("pointerdown", onPointerDown);
    element.addEventListener("pointermove", onPointerMove);
    element.addEventListener("pointerup", onPointerUp);
    element.addEventListener("pointercancel", onPointerUp);
    return {applyTransform, detach, state};
  }

  function openMermaidOverlay(originalSvg) {
    if (overlayCleanup) {
      overlayCleanup();
    }
    if (!overlayEl) {
      overlayEl = document.createElement("div");
      overlayEl.classList.add("mermaid-overlay");
      overlayEl.setAttribute("role", "dialog");
      overlayEl.setAttribute("aria-modal", "true");
      overlayEl.setAttribute("aria-label", "Diagram fullscreen view");
      document.body.appendChild(overlayEl);
    }
    const svgClone = originalSvg.cloneNode(true);
    svgClone.removeAttribute("data-panzoom");
    svgClone.style.transform = "";
    svgClone.style.transformOrigin = "0 0";
    svgClone.style.display = "block";
    const inner = document.createElement("div");
    inner.classList.add("mermaid-overlay-inner");
    inner.appendChild(svgClone);
    const closeBtn = document.createElement("button");
    closeBtn.setAttribute("type", "button");
    closeBtn.classList.add("mermaid-overlay-close");
    closeBtn.setAttribute("aria-label", "Close fullscreen view");
    closeBtn.textContent = "×";
    const resetBtn = document.createElement("button");
    resetBtn.setAttribute("type", "button");
    resetBtn.classList.add("mermaid-reset-btn");
    resetBtn.setAttribute("aria-label", "Reset diagram view");
    resetBtn.textContent = "Reset";
    overlayEl.innerHTML = "";
    overlayEl.appendChild(inner);
    overlayEl.appendChild(closeBtn);
    overlayEl.appendChild(resetBtn);
    const panZoom = attachPanZoom(inner, svgClone);
    document.body.style.overflow = "hidden";
    overlayEl.classList.add("is-open");
    const innerRect = inner.getBoundingClientRect();
    const svgRect = svgClone.getBoundingClientRect();
    if (svgRect.width > 0 && svgRect.height > 0) {
      const fitScale = Math.min(
        1,
        (innerRect.width - mermaidOverlayPadding) / svgRect.width,
        (innerRect.height - mermaidOverlayPadding) / svgRect.height
      );
      panZoom.state.scale = fitScale;
      panZoom.state.tx = (innerRect.width - svgRect.width * fitScale) / 2;
      panZoom.state.ty = (innerRect.height - svgRect.height * fitScale) / 2;
      panZoom.applyTransform();
    }
    const initScale = panZoom.state.scale;
    const initTx = panZoom.state.tx;
    const initTy = panZoom.state.ty;
    const onKeyDown = (e) => {
      if (e.key === "Escape") {
        closeOverlay();
      }
    };
    const closeOverlay = () => {
      overlayEl.classList.remove("is-open");
      document.body.style.overflow = "";
      document.removeEventListener("keydown", onKeyDown);
      panZoom.detach();
      overlayCleanup = null;
    };
    resetBtn.addEventListener("click", () => {
      panZoom.state.scale = initScale;
      panZoom.state.tx = initTx;
      panZoom.state.ty = initTy;
      panZoom.applyTransform();
    });
    closeBtn.addEventListener("click", closeOverlay);
    document.addEventListener("keydown", onKeyDown);
    overlayCleanup = closeOverlay;
  }

  function setupMermaidPanZoom() {
    document.querySelectorAll(mermaidQuery).forEach((element) => {
      const svg = element.querySelector("svg");
      if (!svg || svg.hasAttribute("data-panzoom")) {
        return;
      }
      const pre = element.closest("pre");
      const oldResetBtn = (
        pre
        ? pre.querySelector(".mermaid-reset-btn")
        : null
      );
      const oldExpandBtn = (
        pre
        ? pre.querySelector(".mermaid-expand-btn")
        : null
      );
      const panZoom = attachPanZoom(element, svg);
      const resetBtn = document.createElement("button");
      const expandBtn = document.createElement("button");
      svg.setAttribute("data-panzoom", "true");
      svg.style.transformOrigin = "0 0";
      svg.style.display = "block";
      if (pre) {
        pre.classList.add("diagram-mermaid-pre");
        pre.style.position = "relative";
      }
      element.classList.add("diagram-mermaid-code");
      resetBtn.setAttribute("type", "button");
      resetBtn.classList.add("mermaid-reset-btn");
      resetBtn.setAttribute("aria-label", "Reset diagram view");
      resetBtn.textContent = "Reset";
      resetBtn.addEventListener("click", () => {
        panZoom.state.scale = 1;
        panZoom.state.tx = 0;
        panZoom.state.ty = 0;
        panZoom.applyTransform();
      });
      expandBtn.setAttribute("type", "button");
      expandBtn.classList.add("mermaid-expand-btn");
      expandBtn.setAttribute("aria-label", "Expand diagram to full screen");
      expandBtn.innerHTML = expandIcon;
      expandBtn.addEventListener("click", () => openMermaidOverlay(svg));
      if (oldResetBtn) {
        oldResetBtn.remove();
      }
      if (oldExpandBtn) {
        oldExpandBtn.remove();
      }
      if (pre) {
        pre.appendChild(resetBtn);
        pre.appendChild(expandBtn);
      }
    });
  }

  function saveOriginalData(query) {
    return new Promise((resolve, reject) => {
      try {
        const els = document.querySelectorAll(query);
        els.forEach((element) => {
          const pre = element.closest("pre");
          const originalCode = element.textContent;
          element.setAttribute("data-original-code", originalCode);
          if (pre) {
            pre.setAttribute("data-copy-content", originalCode);
          }
        });
        resolve();
      } catch (error) {
        reject(error);
      }
    });
  }

  function resetProcessed(query) {
    return new Promise((resolve, reject) => {
      try {
        const els = document.querySelectorAll(query);
        els.forEach((element) => {
          if (element.getAttribute("data-original-code") !== null) {
            element.removeAttribute("data-processed");
            element.innerHTML = "";
            element.textContent = element.getAttribute("data-original-code");
          }
        });
        resolve();
      } catch (error) {
        reject(error);
      }
    });
  }

  function getMapFeatures(code, isTopoJSON) {
    const parsed = JSON.parse(code);
    let featureCollection;

    if (isTopoJSON) {
      if (!parsed.objects || Object.keys(parsed.objects).length === 0) {
        throw new Error("TopoJSON does not contain any objects");
      }

      featureCollection = {
        features: [],
        type: "FeatureCollection"
      };

      Object.values(parsed.objects).forEach((objectValue) => {
        const converted = window.topojson.feature(parsed, objectValue);
        if (converted.type === "FeatureCollection") {
          featureCollection.features = featureCollection.features.concat(converted.features);
        } else {
          featureCollection.features.push(converted);
        }
      });
    } else if (parsed.type === "FeatureCollection") {
      featureCollection = parsed;
    } else if (parsed.type === "Feature") {
      featureCollection = {
        features: [parsed],
        type: "FeatureCollection"
      };
    } else {
      featureCollection = {
        features: [{
          geometry: parsed,
          properties: {},
          type: "Feature"
        }],
        type: "FeatureCollection"
      };
    }

    if (!featureCollection.features || featureCollection.features.length === 0) {
      throw new Error("Geo data does not contain any features");
    }

    return featureCollection;
  }

  function renderMap(codeElement) {
    const isTopoJSON = codeElement.matches(topoJSONQuery);
    const pre = codeElement.closest("pre");
    const featureCollection = getMapFeatures(
      codeElement.getAttribute("data-original-code") || codeElement.textContent,
      isTopoJSON
    );
    const mapContainer = document.createElement("div");

    if (pre) {
      pre.classList.add("diagram-map-pre");
    }

    codeElement.classList.add("diagram-map-code");
    mapContainer.classList.add("leaflet-diagram-map");
    codeElement.innerHTML = "";
    codeElement.appendChild(mapContainer);
    codeElement.setAttribute("data-processed", "true");

    const map = window.L.map(mapContainer, {
      attributionControl: true,
      scrollWheelZoom: false,
      zoomControl: true
    });

    window.L.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
      attribution: "&copy; OpenStreetMap contributors",
      maxZoom: 19
    }).addTo(map);

    const layer = window.L.geoJSON(featureCollection, {
      pointToLayer: function (feature, latlng) {
        let pointRadius = 6;

        if (feature && feature.properties) {
          pointRadius = 6;
        }

        return window.L.circleMarker(latlng, {
          color: "#1f6feb",
          fillColor: "#58a6ff",
          fillOpacity: 0.8,
          opacity: 1,
          radius: pointRadius,
          weight: 2
        });
      },
      style: function (feature) {
        const geometryType = (
          (
            feature &&
            feature.geometry &&
            feature.geometry.type
          )
          ? feature.geometry.type
          : ""
        );
        const hasFill = (
          geometryType.indexOf("Polygon") !== -1 ||
          geometryType.indexOf("MultiPolygon") !== -1
        );

        return {
          color: "#1f6feb",
          fillColor: "#58a6ff",
          fillOpacity: (
            hasFill
            ? 0.22
            : 0
          ),
          opacity: 0.9,
          weight: 2
        };
      }
    }).addTo(map);

    if (layer.getBounds && layer.getBounds().isValid()) {
      map.fitBounds(layer.getBounds(), {padding: [24, 24]});
    } else {
      map.setView([20, 0], 1);
    }
  }

  function renderMaps() {
    saveOriginalData(mapQuery).then(() => {
      document.querySelectorAll(mapQuery).forEach((element) => {
        try {
          renderMap(element);
        } catch (error) {
          console.error(error);
        }
      });
    }).catch(console.error);
  }

  async function initMermaid() {
    // Workaround issue with MermaidJS that doesn't allow changing theme on
    // re-initialization
    // https://github.com/mermaid-js/mermaid/issues/1945#issuecomment-1661264708
    saveOriginalData(mermaidQuery).catch(console.error);

    if (window.Param.mode === "dark") {
      await loadMermaid(false);
    } else if (window.Param.mode === "light") {
      await loadMermaid(true);
    } else {
      if (!diagramMediaQuery) {
        diagramMediaQuery = window.matchMedia("(prefers-color-scheme: light)");
      }
      await loadMermaid(diagramMediaQuery.matches);
      if (!diagramMediaQuery.codexListenerAdded) {
        diagramMediaQuery.addEventListener("change", (e) => {
          resetProcessed(mermaidQuery).then(() => loadMermaid(e.matches)).catch(console.error);
        });
        diagramMediaQuery.codexListenerAdded = true;
      }
    }
  }

  async function renderDiagrams() {
    await initMermaid();
    document.querySelectorAll(mapQuery).forEach((element) => {
      element.removeAttribute("data-processed");
    });
    renderMaps();
  }

  async function loadMarkdown() {
    const response = await fetch(`/__/md?path=${window.location.pathname.slice(1)}`);
    const result = await response.json();

    const markdownBody = document.getElementById("markdown-body");
    markdownBody.innerHTML = result.html;

    const markdownTitle = document.getElementById("markdown-title");
    markdownTitle.innerHTML = result.title;

    updateHeadingsList(result.headings_html, result.has_headings);

    await renderDiagrams();
    await typesetMathJax();
    addCopyButtons();
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
      if (pre.querySelector(".copy-button")) {
        return;
      }

      const code = pre.querySelector("code");
      if (!code) {
        return;
      }

      const button = document.createElement("button");
      button.classList.add("copy-button");
      button.setAttribute("aria-label", "Copy code to clipboard");
      button.innerHTML = copyIcon;
      pre.style.position = "relative";
      pre.appendChild(button);

      button.addEventListener("click", () => {
        const copyContent = pre.getAttribute("data-copy-content") || code.textContent;
        navigator.clipboard.writeText(copyContent).then(() => {
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

  function updateHeadingsList(headingsHTML, hasHeadings) {
    const details = document.getElementById("heading-list");
    const list = document.getElementById("headings-tree");

    if (!details || !list) {
      return;
    }

    list.innerHTML = headingsHTML || "";
    if (!hasHeadings) {
      details.classList.add("is-disabled");
      details.setAttribute("aria-disabled", "true");
      details.open = false;
      return;
    }

    details.classList.remove("is-disabled");
    details.removeAttribute("aria-disabled");
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
  const popovers = [];
  const fileBrowser = document.getElementById("file-browser");
  if (fileBrowser) {
    popovers.push(fileBrowser);
  }
  const headingList = document.getElementById("heading-list");
  if (headingList) {
    popovers.push(headingList);
  }

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

  popovers.forEach((details) => {
    const summary = details.querySelector("summary");
    if (!summary) {
      return;
    }
    summary.addEventListener("click", function (e) {
      if (details.classList.contains("is-disabled")) {
        e.preventDefault();
        details.open = false;
      }
    });

    if (details.id !== "heading-list") {
      return;
    }

    const headingsTree = details.querySelector("#headings-tree");
    if (!headingsTree) {
      return;
    }

    headingsTree.addEventListener("click", function (e) {
      if (e.target && e.target.closest && e.target.closest(".heading-item")) {
        details.open = false;
      }
    });
  });
});
