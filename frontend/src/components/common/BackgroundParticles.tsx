import React, { useEffect, useState, useContext } from "react";
import Particles, { initParticlesEngine } from "@tsparticles/react";
import { loadFull } from "tsparticles";
import type { ISourceOptions } from "@tsparticles/engine";
import { ThemeContext } from "../helpers/ThemeContext";

const BackgroundParticles: React.FC = () => {
  const [particlesReady, setParticlesReady] = useState(false);
  const { theme } = useContext(ThemeContext);
  const [options, setOptions] = useState<ISourceOptions | null>(null);

  useEffect(() => {
    initParticlesEngine(async (engine) => {
      await loadFull(engine);
    }).then(() => setParticlesReady(true));
  }, []);

  useEffect(() => {
    if (!particlesReady) return;

    const isDark = theme === "dark";
    const backgroundColor = isDark ? "#0d0d2b" : "#e0e0e0";
    const particleColor = isDark ? "#ffffff" : "#000000";

    const newOptions: ISourceOptions = {
      fullScreen: { enable: true, zIndex: -1 }, // Фон на весь экран, под контентом
      background: { color: { value: backgroundColor } },
      fpsLimit: 60,
      interactivity: {
        events: {
          onHover: { enable: true, mode: "repulse" },
          resize: { enable: true },
        },
        modes: { repulse: { distance: 100, duration: 0.4 } },
      },
      particles: {
        color: { value: particleColor },
        links: {
          color: { value: particleColor },
          distance: 130,
          enable: true,
          opacity: 0.2,
          width: 1,
        },
        move: {
          enable: true,
          speed: 1,
          outModes: { default: "bounce" },
        },
        number: {
          value: 60,
          density: { enable: true, width: 800, height: 800 },
        },
        opacity: { value: 0.5 },
        shape: { type: "circle" },
        size: { value: { min: 1, max: 3 } },
      },
    };

    setOptions(newOptions);
  }, [theme, particlesReady]);

  return (
    <>
      {particlesReady && options && (
        <Particles
          id="tsparticles"
          key={theme}
          options={options}
          style={{ pointerEvents: "none" }} // ✅ чтобы не мешал кликам
        />
      )}
    </>
  );
};

export default BackgroundParticles;
