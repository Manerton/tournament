import React, { useState } from "react";
import { Navbar, Nav, Dropdown, Container, Button } from "react-bootstrap";
import { LinkContainer } from "react-router-bootstrap";
import { Link } from "react-router-dom";
import { PersonCircle, BoxArrowInRight, ChevronUp, ChevronDown } from "react-bootstrap-icons";
import ThemeToggleButton from "../helpers/ThemeToggleButton";
import { useAuth } from "../helpers/AuthContext";

function Header() {
  //const { user, logout, initialized } = useAuth();
  const [visible, setVisible] = useState(true); // состояние для скрытия/показа Navbar
    const initialized=true;
  if (!initialized) {
    return (
      <Navbar className="bg-body-tertiary border border-1">
        <Container>Загрузка...</Container>
      </Navbar>
    );
  }

  const isAuthenticated = true;

  if (!visible) {
    return (
      <div className="text-center mt-2">
        <Button
          variant="outline-secondary"
          size="sm"
          onClick={() => setVisible(true)}
          className="d-inline-flex align-items-center"
        >
          <ChevronDown className="me-1" /> Показать панель
        </Button>
      </div>
    );
  }

  return (
    <>
      <Navbar expand="lg" sticky="top" className="bg-body-tertiary border border-1 shadow-sm">
        <Container fluid="lg" className="mx-auto">
          <Navbar.Brand as={Link} to="/" className="d-flex align-items-center">
            <img
                src="/favicon.png"
                alt="Логотип"
                width="32"
                height="32"
                className="d-inline-block align-top me-2"
            />
            ВСОШ
                </Navbar.Brand>

          <Navbar.Toggle aria-controls="main-navbar" />
          <Navbar.Collapse id="main-navbar">
            <Nav className="me-auto">
              <LinkContainer to="/tournaments">
                <Nav.Link>Турниры</Nav.Link>
              </LinkContainer>

              {isAuthenticated && (
                <LinkContainer to="/admin">
                  <Nav.Link>Админ панель</Nav.Link>
                </LinkContainer>
              )}
            </Nav>

            <Nav className="ms-auto align-items-center">
              <ThemeToggleButton />

              {isAuthenticated ? (
                <Dropdown align="end" className="ms-3">
                  <Dropdown.Toggle
                    variant="light"
                    id="user-dropdown"
                    className="d-flex align-items-center border-0 bg-transparent"
                  >
                    <PersonCircle size={24} className="me-2" />
                    <span className="text-body">
                      {/* {user?.email || "Пользователь"} */}
                    </span>
                  </Dropdown.Toggle>

                  <Dropdown.Menu>
                    <Dropdown.Item as={Link} to="/profile">
                      <PersonCircle className="me-2" /> Мой профиль
                    </Dropdown.Item>
                    {/* <Dropdown.Item onClick={logout}>
                      <BoxArrowInRight className="me-2" /> Выйти
                    </Dropdown.Item> */}
                  </Dropdown.Menu>
                </Dropdown>
              ) : (
                <Nav.Link as={Link} to="/login" className="ms-3 d-flex align-items-center">
                  <BoxArrowInRight className="me-2" />
                  Войти
                </Nav.Link>
              )}
            </Nav>
          </Navbar.Collapse>
        </Container>
      </Navbar>

      {/* Кнопка для сворачивания Navbar */}
      <div className="text-center mt-2">
        <Button
          variant="outline-secondary"
          size="sm"
          onClick={() => setVisible(false)}
          className="d-inline-flex align-items-center"
        >
          <ChevronUp className="me-1" /> Скрыть панель
        </Button>
      </div>
    </>
  );
}

export default Header;
