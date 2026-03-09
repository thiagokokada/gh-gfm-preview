import * as THREE from "/static/generated/three.module.js";
import { OrbitControls } from "/static/generated/stl/OrbitControls.js";
import { STLLoader } from "/static/generated/stl/STLLoader.js";

const stlQuery = "code.language-stl";
const modeList = ["solid", "wireframe", "xray"];
const loader = new STLLoader();

function isLightMode() {
  if (window.Param && window.Param.mode === "light") {
    return true;
  }
  if (window.Param && window.Param.mode === "dark") {
    return false;
  }
  return window.matchMedia("(prefers-color-scheme: light)").matches;
}

function getSource(codeElement) {
  const source = codeElement.textContent;
  codeElement.setAttribute("data-original-code", source);
  if (codeElement.parentElement) {
    codeElement.parentElement.setAttribute("data-copy-content", source);
  }
  return source;
}

function applyMaterialMode(mesh, mode) {
  if (mode === "wireframe") {
    mesh.material.color.set("#58a6ff");
    mesh.material.opacity = 1;
    mesh.material.transparent = false;
    mesh.material.wireframe = true;
    return;
  }

  mesh.material.wireframe = false;
  if (mode === "xray") {
    mesh.material.color.set("#79c0ff");
    mesh.material.opacity = 0.38;
    mesh.material.transparent = true;
    return;
  }

  mesh.material.color.set("#58a6ff");
  mesh.material.opacity = 1;
  mesh.material.transparent = false;
}

function fitCamera(camera, controls, geometry) {
  geometry.computeBoundingBox();
  geometry.computeBoundingSphere();

  const center = geometry.boundingSphere.center.clone();
  const radius = Math.max(geometry.boundingSphere.radius, 1);
  const distance = radius * 2.8;

  camera.near = Math.max(radius / 100, 0.1);
  camera.far = radius * 20;
  camera.position.set(center.x + distance, center.y + distance, center.z + distance);
  camera.lookAt(center);
  camera.updateProjectionMatrix();
  controls.target.copy(center);
  controls.update();
}

function createButton(label, mode, onClick) {
  const button = document.createElement("button");
  button.classList.add("stl-diagram-button");
  button.textContent = label;
  button.type = "button";
  button.addEventListener("click", () => onClick(mode));
  return button;
}

function renderSTL(codeElement) {
  if (codeElement.getAttribute("data-processed") === "true") {
    return;
  }

  const pre = codeElement.closest("pre");
  const source = getSource(codeElement);
  const geometry = loader.parse(new TextEncoder().encode(source).buffer);
  const lightMode = isLightMode();
  const wrapper = document.createElement("div");
  const toolbar = document.createElement("div");
  const modes = document.createElement("div");
  const hint = document.createElement("div");
  const viewport = document.createElement("div");
  const scene = new THREE.Scene();
  const camera = new THREE.PerspectiveCamera(40, 1, 0.1, 1000);
  const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
  const controls = new OrbitControls(camera, renderer.domElement);
  const material = new THREE.MeshPhongMaterial({
    color: 0x58a6ff,
    flatShading: true,
  });
  const mesh = new THREE.Mesh(geometry, material);
  const ambientLight = new THREE.AmbientLight((lightMode ? 0xffffff : 0x8b949e), 1.8);
  const mainLight = new THREE.DirectionalLight(0xffffff, 2.4);
  const rimLight = new THREE.DirectionalLight((lightMode ? 0x8b949e : 0x79c0ff), 1.4);
  const buttons = {};
  let animationId = 0;

  wrapper.classList.add("stl-diagram");
  toolbar.classList.add("stl-diagram-toolbar");
  modes.classList.add("stl-diagram-modes");
  hint.classList.add("stl-diagram-hint");
  viewport.classList.add("stl-diagram-viewport");
  hint.textContent = "Drag to orbit, scroll to zoom";

  if (pre) {
    pre.classList.add("diagram-stl-pre");
  }
  codeElement.classList.add("diagram-stl-code");
  codeElement.innerHTML = "";
  toolbar.appendChild(modes);
  toolbar.appendChild(hint);
  wrapper.appendChild(toolbar);
  wrapper.appendChild(viewport);
  codeElement.appendChild(wrapper);
  codeElement.setAttribute("data-processed", "true");

  renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));
  renderer.setClearColor((lightMode ? 0xf6f8fa : 0x010409), 1);
  viewport.appendChild(renderer.domElement);

  controls.enableDamping = true;
  controls.dampingFactor = 0.08;
  controls.minDistance = 1;
  controls.maxDistance = 1000;

  geometry.center();
  geometry.computeVertexNormals();

  scene.add(mesh);
  scene.add(ambientLight);
  mainLight.position.set(6, 8, 10);
  rimLight.position.set(-4, -3, -7);
  scene.add(mainLight);
  scene.add(rimLight);

  fitCamera(camera, controls, geometry);
  applyMaterialMode(mesh, "solid");

  [
    ["Solid", "solid"],
    ["Wireframe", "wireframe"],
    ["X-Ray", "xray"],
  ].forEach(([label, mode]) => {
    const button = createButton(label, mode, (selectedMode) => {
      applyMaterialMode(mesh, selectedMode);
      modeList.forEach((currentMode) => {
        buttons[currentMode].classList.toggle("is-active", currentMode === selectedMode);
      });
    });
    buttons[mode] = button;
    modes.appendChild(button);
  });
  buttons.solid.classList.add("is-active");

  function resize() {
    const width = viewport.clientWidth || 640;
    const height = viewport.clientHeight || 400;

    camera.aspect = width / height;
    camera.updateProjectionMatrix();
    renderer.setSize(width, height, false);
  }

  function animate() {
    if (!document.body.contains(viewport)) {
      renderer.dispose();
      controls.dispose();
      if (animationId) {
        cancelAnimationFrame(animationId);
      }
      return;
    }

    controls.update();
    renderer.render(scene, camera);
    animationId = requestAnimationFrame(animate);
  }

  resize();
  if (window.ResizeObserver) {
    const observer = new ResizeObserver(() => resize());
    observer.observe(viewport);
  } else {
    window.addEventListener("resize", resize);
  }
  animate();
}

function renderSTLDiagrams() {
  document.querySelectorAll(stlQuery).forEach((codeElement) => {
    try {
      renderSTL(codeElement);
    } catch (error) {
      console.error(error);
    }
  });
}

window.renderSTLDiagrams = renderSTLDiagrams;
renderSTLDiagrams();
