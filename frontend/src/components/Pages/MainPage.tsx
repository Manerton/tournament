// src/pages/MainPage.tsx
import React from "react";
import { Container } from "react-bootstrap";
import "../../assets/MainPage.css";

import MainLogo from "/battleroyale main.png";

const MainPage: React.FC = () => {
  return (
    <div className="mainpage-content">
      <Container>
        {/* <h1 className="fade-in display-3 fw-bold mb-3 text">
          Турнирная система РШТ
          <p className="fade-in delay-1 lead text">
          </h1>
          Место, где закаляются умы
        </p> */}
        <img src={MainLogo} alt="Турнирная система РШТ" className="w-50  h-150" />
      </Container>
    </div>
  );
};

export default MainPage;
