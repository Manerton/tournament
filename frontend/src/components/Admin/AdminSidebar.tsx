import { useState } from "react";
import { Nav, Button } from "react-bootstrap";
import { Link } from "react-router-dom";
import {
  HouseDoor,
  People,
  FileEarmarkText,
  CreditCard,
  CalendarCheck,
  FileText,
  BoxArrowLeft,
  PersonCircle,
} from "react-bootstrap-icons";

import ThemeToggleButton from "../helpers//ThemeToggleButton";
import { useAuth } from "../helpers/AuthContext";

const AdminSidebar: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false);
  const { user, logout } = useAuth();

  const toggleSidebar = () => setCollapsed(!collapsed);

  return (
    <div
      className={`d-flex flex-column justify-content-between bg-dark text-white border-end ${
        collapsed ? "px-2" : "px-3"
      }`}
      style={{
        width: collapsed ? "70px" : "250px",
        height: "100vh",
        position: "sticky",
        top: 0,
        zIndex: 1020,
        transition: "width 0.3s",
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
                <div className="small">
                  id-{user.id}
                </div>
              </div>
            )}
          </div>
          <Button
            variant="link"
            className="text-white p-0 ms-auto"
            onClick={toggleSidebar}
          >
            <span style={{ fontSize: "1.2rem" }}>{collapsed ? "»" : "«"}</span>
          </Button>
        </div>

        <hr className="border-secondary" />

        {/* Навигация */}
        <Nav className="flex-column gap-2">
          <div className="text-uppercase small px-2">
            {!collapsed && "Разделы"}
          </div>

          <Nav.Link
            as={Link}
            to="/"
            className="text-white d-flex align-items-center"
          >
            <HouseDoor className="me-2" />
            {!collapsed && "На главную"}
          </Nav.Link>

          {/*<Nav.Link*/}
          {/*  as={Link}*/}
          {/*  to={`/applications/user/${user?.id}`}*/}
          {/*  className="text-white d-flex align-items-center"*/}
          {/*>*/}
          {/*  <FileText className="me-2" />*/}
          {/*  {!collapsed && "Статус заявки"}*/}
          {/*</Nav.Link>*/}

          {/* Админские ссылки */}
          {user && (
            <>

                <Nav.Link
                as={Link}
                to="/TeamsApplications"
                className="text-white d-flex align-items-center"
              >
                <FileText className="me-2" />
                {!collapsed && "Заявки команд"}
              </Nav.Link>

              <Nav.Link
                as={Link}
                to="/participants"
                className="text-white d-flex align-items-center"
              >
                <People className="me-2" />
                {!collapsed && "Пользователи"}
              </Nav.Link>

              <Nav.Link
                as={Link}
                to="/olymp-admin/event/index"
                className="text-white d-flex align-items-center"
              >
                <CalendarCheck className="me-2" />
                {!collapsed && "Турниры"}
              </Nav.Link>

              <Nav.Link
                as={Link}
                to="/schools"
                className="text-white d-flex align-items-center"
              >
                <FileEarmarkText className="me-2" />
                {!collapsed && "Школы"}
              </Nav.Link>

              <Nav.Link
                as={Link}
                to="/report"
                className="text-white d-flex align-items-center"
              >
                <CreditCard className="me-2" />
                {!collapsed && "Отчеты"}
              </Nav.Link>

            
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
  );
};

export default AdminSidebar;
