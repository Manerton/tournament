import { useState } from "react";
import { Navbar, Nav, Dropdown, Container } from "react-bootstrap";
import { LinkContainer } from "react-router-bootstrap";
import { Link } from "react-router-dom";
import {
  PersonCircle,
  BoxArrowInRight,
  ChevronUp,
  ChevronDown,
} from "react-bootstrap-icons";
import ThemeToggleButton from "../helpers/ThemeToggleButton";
import { useAuth } from "../helpers/AuthContext";

function Header() {
  const { user, logout, initialized } = useAuth();
  const [collapsed, setCollapsed] = useState(false);

  if (!initialized) {
    return (
      <Navbar className="bg-body-tertiary border border-1">
        <Container>Загрузка...</Container>
      </Navbar>
    );
  }

  const isAuthenticated = !!user;

  return (
    <div
      className="header-container border-bottom border-1 shadow-sm dynamic-bg"
      style={{
        position: "sticky",
        top: 0,
        zIndex: 1020,
        transition: "height 0.3s",
        overflow: "hidden",
      }}
    >
      {/* Основная панель навигации */}
      <div
        style={{
          height: collapsed ? "0" : "auto",
          overflow: "hidden",
          transition: "height 0.3s ease",
        }}
      >
        <Navbar expand="lg" className="rounded-0">
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
                      <Dropdown.Item as={Link} to="/login">
                        <PersonCircle className="me-2" /> Профиль
                      </Dropdown.Item>
                      <Dropdown.Item onClick={logout}>
                        <BoxArrowInRight className="me-2" /> Выйти
                      </Dropdown.Item>
                    </Dropdown.Menu>
                  </Dropdown>
                ) : (
                  <Nav.Link
                    as={Link}
                    to="/login"
                    className="ms-3 d-flex align-items-center"
                  >
                    <BoxArrowInRight className="me-2" />
                    Войти
                  </Nav.Link>
                )}
              </Nav>
            </Navbar.Collapse>
          </Container>
        </Navbar>
      </div>

      {/* Нижняя “полоса” — кнопка сворачивания/разворачивания */}
      <div
        className="navbar-toggle d-flex justify-content-center align-items-center border-top border-secondary border-opacity-25"
        onClick={() => setCollapsed(!collapsed)}
        style={{
          height: "2rem",
          cursor: "pointer",
          userSelect: "none",
          transition: "background-color 0.2s",
        }}
        title={collapsed ? "Развернуть панель" : "Свернуть панель"}
      >
        <span
          style={{
            fontSize: "1.5rem",
            color: "var(--bs-body-color)",
            lineHeight: 1,
          }}
        >
          {collapsed ? <ChevronDown /> : <ChevronUp />}
        </span>
      </div>
    </div>
  );
}

export default Header;
