import { useState } from "react";
import { Navbar, Nav, Dropdown, Container, Button } from "react-bootstrap";
import { LinkContainer } from "react-router-bootstrap";
import { Link } from "react-router-dom";
import { PersonCircle, BoxArrowInRight, ChevronUp, ChevronDown } from "react-bootstrap-icons";
import ThemeToggleButton from "../helpers/ThemeToggleButton";
import { useAuth } from "../helpers/AuthContext";
import { Background } from "@tsparticles/engine";

function Header() {
  const { user, logout, initialized } = useAuth();
  const [visible, setVisible] = useState(true);

  if (!initialized) {
    return (
      <Navbar className="bg-body-tertiary border border-1">
        <Container>Загрузка...</Container>
      </Navbar>
    );
  }

  const isAuthenticated = !!user; 

  if (!visible) {
    return (
      <div className="text-center mt-2">
        <Button
          variant="outline-secondary"
          size="sm"
          onClick={() => setVisible(true)}
          className="d-inline-flex align-items-center"
          aria-label="Показать панель"
        >
          <ChevronDown className="me-1" /> Показать панель
        </Button>
      </div>
    );
  }

  return (
    <>
      <Navbar expand="lg" sticky="top" className="border border-1 shadow-sm rounded dynamic-bg">
        <Container fluid="lg" className="mx-auto">
          <Navbar.Brand as={Link} to="/" className="d-flex align-items-center">
            <img
              src="/favicon.png"
              height={32}
              width={80}
              alt="Логотип"
              className="d-inline-block align-top me-2"
            />
          </Navbar.Brand>

          <Navbar.Toggle aria-controls="main-navbar" />
          <Navbar.Collapse id="main-navbar">
            <Nav className="me-auto">
              <LinkContainer to="/tournaments">
                <Nav.Link>Турниры</Nav.Link>
              </LinkContainer>

              {isAuthenticated && (
                <LinkContainer to="/AdminPanel">
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
                      {user?.email || user?.id}
                    </span>
                  </Dropdown.Toggle>

                  <Dropdown.Menu>
                    { }
                    <Dropdown.Item as={Link} to="/login">
                      <PersonCircle className="me-2" /> Профиль
                    </Dropdown.Item>
                    <Dropdown.Item onClick={logout}>
                      <BoxArrowInRight className="me-2" /> Выйти
                    </Dropdown.Item>
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

      <div className="text-center mt-2">
        <Button
          variant="outline-secondary"
          size="sm"
          onClick={() => setVisible(false)}
          className="d-inline-flex align-items-center"
          aria-label="Скрыть панель"
        >
          <ChevronUp className="me-1" /> Скрыть панель
        </Button>
      </div>
    </>
  );
}

export default Header;