// MainPage.tsx
import React, { useEffect, useState } from "react";
import { Container } from "react-bootstrap";
import Particles, { initParticlesEngine } from "@tsparticles/react";
import { loadFull } from "tsparticles";
import type { ISourceOptions } from "@tsparticles/engine";
import "../../assets/MainPage.css";

const MainPage: React.FC = () => {
  const [particlesReady, setParticlesReady] = useState(false);

  useEffect(() => {
    initParticlesEngine(async (engine) => {
      await loadFull(engine);
    }).then(() => {
      setParticlesReady(true);
    });
  }, []);

  const options: ISourceOptions = {
    fullScreen: {
      enable: true,
      zIndex: 0,
    },
    background: {
      color: { value: "#0d0d2b" },
    },
    fpsLimit: 60,
    interactivity: {
      events: {
        onHover: { enable: true, mode: "repulse" },
        onClick: { enable: false },
        resize: { enable: true },
      },
      modes: {
        repulse: { distance: 100, duration: 0.4 },
      },
    },
    particles: {
      color: { value: "#ffffff" },
      links: {
        color: { value: "#ffffff" },
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
        density: {
          enable: true,
          width: 800,
          height: 800,
        },
      },
      opacity: { value: 0.5 },
      shape: { type: "circle" },
      size: { value: { min: 1, max: 3 } },
    },
  };

  return (
    <div className="mainpage-wrapper">
      {particlesReady && <Particles id="tsparticles" options={options} />}
      <div className="mainpage-content text-center">
        <Container>
          <h1 className="fade-in display-3 fw-bold mb-3 text-light">
            Турнирная система РШТ
          </h1>
          <p className="fade-in delay-1 lead text-light">
            Место, где закаляются умы
          </p>
        </Container>
      </div>
    </div>
  );
};

export default MainPage;
