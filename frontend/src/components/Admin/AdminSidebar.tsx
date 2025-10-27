import { useState } from "react";
import { Nav } from "react-bootstrap";
import { LinkContainer } from "react-router-bootstrap";
import {
  HouseDoor,
  People,
  FileEarmarkText,
  CreditCard,
  CalendarCheck,
  FileText,
  BoxArrowLeft,
  PersonCircle,
  ChevronRight,
  ChevronLeft,
} from "react-bootstrap-icons";
import ThemeToggleButton from "../helpers/ThemeToggleButton";
import { useAuth } from "../helpers/AuthContext";

const AdminSidebar: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false);
  const { user, logout } = useAuth();
  const toggleSidebar = () => setCollapsed(!collapsed);

  return (
    <div
      className="sidebar d-flex border-end shadow-sm dynamic-bg-sidebar text"
      style={{
        height: "100vh",
        position: "sticky",
        top: 0,
        zIndex: 1020,
        transition: "width 0.3s",
      }}
    >
      {/* Левая часть — основное содержимое */}
      <div
        className={`sidebar-main d-flex flex-column justify-content-between ${
          collapsed ? "px-2" : "px-3"
        }`}
        style={{
          width: collapsed ? "90px" : "270px",
          transition: "width 0.3s",
          flexShrink: 0, // не сжимается 
        }}
      >
        {/* Верхняя часть: профиль */}
        <div>
          <div className="d-flex align-items-center justify-content-between py-3">
            <div className="d-flex align-items-center">
              <PersonCircle size={24} className="me-2" />
              {!collapsed && user && (
                <div>
                  <strong>{user.email}</strong>
                  <div className="small">id-{user.id}</div>
                </div>
              )}
            </div>
          </div>

          <hr className="border-secondary" />

          {/* Навигация */}
          <Nav className="flex-column gap-2">
            <div className="text-uppercase small px-2">
              {!collapsed && "Разделы"}
            </div>

            <LinkContainer to="/">
              <Nav.Link className="d-flex align-items-center text">
                <HouseDoor className="me-2" />
                {!collapsed && "На главную"}
              </Nav.Link>
            </LinkContainer>

            {user && (
              <>
                <LinkContainer to="/TeamsApplications">
                  <Nav.Link className="d-flex align-items-center text">
                    <FileText className="me-2" />
                    {!collapsed && "Заявки команд"}
                  </Nav.Link>
                </LinkContainer>

                <LinkContainer to="/participants">
                  <Nav.Link className="d-flex align-items-center text">
                    <People className="me-2" />
                    {!collapsed && "Пользователи"}
                  </Nav.Link>
                </LinkContainer>

                <LinkContainer to="/olymp-admin/event/index">
                  <Nav.Link className="d-flex align-items-center text">
                    <CalendarCheck className="me-2" />
                    {!collapsed && "Турниры"}
                  </Nav.Link>
                </LinkContainer>

                <LinkContainer to="/schools">
                  <Nav.Link className="d-flex align-items-center text">
                    <FileEarmarkText className="me-2" />
                    {!collapsed && "Школы"}
                  </Nav.Link>
                </LinkContainer>

                <LinkContainer to="/report">
                  <Nav.Link className="d-flex align-items-center text">
                    <CreditCard className="me-2" />
                    {!collapsed && "Отчеты"}
                  </Nav.Link>
                </LinkContainer>
              </>
            )}
          </Nav>
        </div>

        <div className="d-flex flex-column pb-3 justify-content-center align-items-start">
          <div className="mb-2">
            <ThemeToggleButton />
          </div>

          <Nav.Link
            onClick={logout}
            className="text-danger align-items-center px-2"
          >
            <BoxArrowLeft className="me-2" />
            {!collapsed && "Выйти"}
          </Nav.Link>
        </div>
      </div>

      {/* Правая колонка — кнопка сворачивания */}
      <div
        className="sidebar-toggle d-flex justify-content-center align-items-center border-start border-secondary border-opacity-50"
        onClick={toggleSidebar}
        style={{
          width: "2rem",
          cursor: "pointer",
          userSelect: "none",
          transition: "background-color 0.2s",
          flexShrink: 0,
        }}
        title={collapsed ? "Развернуть панель" : "Свернуть панель"}
      >
        <span
          style={{
            fontSize: "1.8rem",
            lineHeight: "1",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            color: "var(--bs-body-color)",
          }}
        >
          {collapsed ? <ChevronRight /> : <ChevronLeft />}
        </span>
      </div>
    </div>
  );
};

export default AdminSidebar;
