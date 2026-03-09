/*jslint browser,long,fart,indent2*/
/*global alert,console,window,WebSocket,JSON*/

(function () {
  "use strict";

  const mermaidQuery = "code.language-mermaid";
  const geoJSONQuery = "code.language-geojson";
  const topoJSONQuery = "code.language-topojson";
  const mapQuery = `${geoJSONQuery}, ${topoJSONQuery}`;
  const diagramQuery = `${mermaidQuery}, ${mapQuery}`;
  const diagramWidth = 720;
  const diagramHeight = 420;
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

  function getIsLightMode() {
    if (window.Param.mode === "dark") {
      return false;
    }

    if (window.Param.mode === "light") {
      return true;
    }

    if (!diagramMediaQuery) {
      diagramMediaQuery = window.matchMedia("(prefers-color-scheme: light)");
    }

    return diagramMediaQuery.matches;
  }

  function getMapFeatures(code, isTopoJSON) {
    const parsed = JSON.parse(code);
    let featureCollection;

    if (isTopoJSON) {
      if (!parsed.objects || Object.keys(parsed.objects).length === 0) {
        throw new Error("TopoJSON does not contain any objects");
      }

      featureCollection = {
        type: "FeatureCollection",
        features: [],
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
        type: "FeatureCollection",
        features: [parsed],
      };
    } else {
      featureCollection = {
        type: "FeatureCollection",
        features: [{
          type: "Feature",
          properties: {},
          geometry: parsed,
        }],
      };
    }

    if (!featureCollection.features || featureCollection.features.length === 0) {
      throw new Error("Geo data does not contain any features");
    }

    return featureCollection;
  }

  function renderMap(codeElement, isLight) {
    const isTopoJSON = codeElement.matches(topoJSONQuery);
    const featureCollection = getMapFeatures(
      codeElement.getAttribute("data-original-code") || codeElement.textContent,
      isTopoJSON,
    );
    const projection = window.d3.geoMercator();
    const path = window.d3.geoPath(projection);
    const margin = 16;
    projection.fitExtent(
      [
        [margin, margin],
        [diagramWidth - margin, diagramHeight - margin],
      ],
      featureCollection,
    );

    const backgroundColor = (isLight ? "#f6f8fa" : "#0d1117");
    const strokeColor = (isLight ? "#57606a" : "#8b949e");
    const fillColor = (isLight ? "#dbeafe" : "#1f6feb");
    const emptyFillColor = (isLight ? "#f6f8fa" : "#161b22");

    const svg = window.d3.create("svg")
      .attr("xmlns", "http://www.w3.org/2000/svg")
      .attr("viewBox", `0 0 ${diagramWidth} ${diagramHeight}`)
      .attr("width", "100%")
      .attr("role", "img")
      .attr("aria-label", isTopoJSON ? "Rendered TopoJSON diagram" : "Rendered GeoJSON diagram")
      .style("display", "block")
      .style("max-width", `${diagramWidth}px`)
      .style("height", "auto")
      .style("background", backgroundColor)
      .style("border-radius", "6px");

    svg.append("rect")
      .attr("width", diagramWidth)
      .attr("height", diagramHeight)
      .attr("fill", backgroundColor);

    svg.append("g")
      .selectAll("path")
      .data(featureCollection.features)
      .join("path")
      .attr("d", path)
      .attr("fill", (feature) => {
        if (!feature.geometry || feature.geometry.type.indexOf("Line") !== -1) {
          return "none";
        }
        return fillColor;
      })
      .attr("stroke", strokeColor)
      .attr("stroke-width", 1)
      .attr("vector-effect", "non-scaling-stroke")
      .attr("opacity", 0.95);

    svg.append("path")
      .datum({type: "Sphere"})
      .attr("d", path)
      .attr("fill", "none")
      .attr("stroke", emptyFillColor)
      .attr("stroke-width", 1)
      .attr("vector-effect", "non-scaling-stroke");

    codeElement.innerHTML = "";
    codeElement.appendChild(svg.node());
    codeElement.setAttribute("data-processed", "true");
  }

  function renderMaps(isLight) {
    document.querySelectorAll(mapQuery).forEach((element) => {
      try {
        renderMap(element, isLight);
      } catch (error) {
        console.error(error);
      }
    });
  }

  function renderDiagrams() {
    const isLight = getIsLightMode();

    // Workaround issue with MermaidJS that doesn't allow changing theme on
    // re-initialization
    // https://github.com/mermaid-js/mermaid/issues/1945#issuecomment-1661264708
    saveOriginalData(diagramQuery).then(() => {
      loadMermaid(isLight);
      renderMaps(isLight);
    }).catch(console.error);

    if (window.Param.mode === "auto") {
      if (!diagramMediaQuery) {
        diagramMediaQuery = window.matchMedia("(prefers-color-scheme: light)");
      }

      if (!diagramMediaQuery.codexListenerAdded) {
        diagramMediaQuery.addEventListener("change", () => {
          resetProcessed(diagramQuery).then(() => {
            loadMermaid(getIsLightMode());
            renderMaps(getIsLightMode());
          }).catch(console.error);
        });
        diagramMediaQuery.codexListenerAdded = true;
      }
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
