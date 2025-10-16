import React from "react";
import { Container, Row, Col } from "react-bootstrap";

import logoRHT from "/favicon.png";
import iconVK from "/vk.svg";
import iconOK from "/OK.svg";
import iconTelegram from "/telegram.svg";

const Footer: React.FC = () => (
  <footer className="border-top bg-body-tertiary">
    <Container fluid="lg" className="py-2">
      <Row className="align-items-center gy-3 text-center text-md-start">
        
        {/* Контакты */}
        <Col md={4} className="small text-muted">
          <div>
            <strong>г. Астрахань</strong><br />
            ул. Анри Барбюса, 7<br />
            Тел.: <a href="tel:+78512442428" className="text-decoration-none text-reset">
              +7 8512 442-428
            </a><br />
            <a href="mailto:schooltech@astrobI.ru" className="text-decoration-none text-reset">
              schooltech@astrobI.ru
            </a>
          </div>
        </Col>

         {/* Логотип */}
            <Col md={4} className="d-flex justify-content-center justify-content-md-center">
            <img src={logoRHT} alt="РШТ" style={{ height: "35px" }} />
            </Col>


        {/* Соцсети */}
        <Col md={4} className="d-flex justify-content-center justify-content-md-end gap-4">
          <a href="#" aria-label="ВКонтакте" className="opacity-75 hover-opacity-100">
            <img src={iconVK} alt="VK" style={{ height: "26px" }} />
          </a>
          <a href="#" aria-label="Школьная сеть" className="opacity-75 hover-opacity-100">
            <img src={iconOK} alt="OK" style={{ height: "26px" }} />
          </a>
          <a href="#" aria-label="Telegram" className="opacity-75 hover-opacity-100">
            <img src={iconTelegram} alt="Telegram" style={{ height: "26px" }} />
          </a>
        </Col>
      </Row>
    </Container>

    <div className="bg-secondary text-center py-2 text-white small">
      © 2025 Все права защищены
    </div>
  </footer>
);

export default Footer;
