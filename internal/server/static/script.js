/*jslint browser,long,fart,indent2*/
/*global alert,console,window,WebSocket,JSON*/

(function () {
  "use strict";

  const mermaidQuery = "code.language-mermaid";
  const geoJSONQuery = "code.language-geojson";
  const topoJSONQuery = "code.language-topojson";
  const mapQuery = `${geoJSONQuery}, ${topoJSONQuery}`;
  const copyIcon = `<svg class="copy-icon" aria-hidden="true" fill="none" height="18" shape-rendering="geometricPrecision" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" viewBox="0 0 24 24" width="18" style="color:"currentColor";"><path d="M8 17.929H6c-1.105 0-2-.912-2-2.036V5.036C4 3.91 4.895 3 6 3h8c1.105 0 2 .911 2 2.036v1.866m-6 .17h8c1.105 0 2 .91 2 2.035v10.857C20 21.09 19.105 22 18 22h-8c-1.105 0-2-.911-2-2.036V9.107c0-1.124.895-2.036 2-2.036z"></path></svg>`;
  const tickIcon = `<svg class="tick-icon" aria-hidden="true" fill="none" height="18" shape-rendering="geometricPrecision" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" viewBox="0 0 24 24" width="18" style="color: "currentColor";"><path d="M5 13l4 4L19 7"></path></svg>`;
  let diagramMediaQuery;

  function loadMermaid(isLight) {
    const theme = (
      isLight
      ? "default"
      : "dark"
    );
    window.mermaid.initialize({startOnLoad: false, theme});
    window.mermaid.run({querySelector: mermaidQuery});
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

  function initMermaid() {
    // Workaround issue with MermaidJS that doesn't allow changing theme on
    // re-initialization
    // https://github.com/mermaid-js/mermaid/issues/1945#issuecomment-1661264708
    saveOriginalData(mermaidQuery).catch(console.error);

    if (window.Param.mode === "dark") {
      loadMermaid(false);
    } else if (window.Param.mode === "light") {
      loadMermaid(true);
    } else {
      if (!diagramMediaQuery) {
        diagramMediaQuery = window.matchMedia("(prefers-color-scheme: light)");
      }
      loadMermaid(diagramMediaQuery.matches);
      if (!diagramMediaQuery.codexListenerAdded) {
        diagramMediaQuery.addEventListener("change", (e) => {
          resetProcessed(mermaidQuery).then(() => {
            loadMermaid(e.matches);
          }).catch(console.error);
        });
        diagramMediaQuery.codexListenerAdded = true;
      }
    }
  }

  function renderDiagrams() {
    initMermaid();
    document.querySelectorAll(mapQuery).forEach((element) => {
      element.removeAttribute("data-processed");
    });
    renderMaps();
    if (window.renderSTLDiagrams) {
      window.renderSTLDiagrams();
    }
  }

  async function loadMarkdown() {
    const response = await fetch(`/__/md?path=${window.location.pathname.slice(1)}`);
    const result = await response.json();

    const markdownBody = document.getElementById("markdown-body");
    markdownBody.innerHTML = result.html;

    const markdownTitle = document.getElementById("markdown-title");
    markdownTitle.innerHTML = result.title;

    updateHeadingsList(result.headings_html, result.has_headings);

    renderDiagrams();
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
