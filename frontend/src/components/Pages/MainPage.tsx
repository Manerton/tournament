// src/pages/MainPage.tsx
import React from "react";
import { Container } from "react-bootstrap";
import "../../assets/MainPage.css";

const MainPage: React.FC = () => {
  return (
    <div className="mainpage-content text-center">
      <Container>
        <h1 className="fade-in display-3 fw-bold mb-3 text">
          Турнирная система РШТ
        </h1>
        <p className="fade-in delay-1 lead text">
          Место, где закаляются умы
        </p>
      </Container>
    </div>
  );
};

export default MainPage;
