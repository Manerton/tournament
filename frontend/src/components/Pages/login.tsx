import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../helpers/AuthContext";
import { Form, Button, Alert, Card, Container, Spinner } from "react-bootstrap";

const LoginPage: React.FC = () => {
  const { login, user, initialized } = useAuth();
  const navigate = useNavigate();

  const [formLogin, setFormLogin] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  // Если уже вошли — редирект
  if (initialized && user) {
    navigate("/");
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      await login(formLogin, password);
      navigate("/"); // редирект после успеха
    } catch (err: any) {
      console.error("Login failed:", err);
      setError("Неверный логин или пароль");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container className="d-flex align-items-center justify-content-center vh-100">
      <Card className="p-4 shadow-lg" style={{ minWidth: "320px", maxWidth: "400px" }}>
        <h3 className="text-center mb-4">Авторизация</h3>

        {error && <Alert variant="danger">{error}</Alert>}

        <Form onSubmit={handleSubmit}>
          <Form.Group controlId="login" className="mb-3">
            <Form.Label>Логин</Form.Label>
            <Form.Control
              type="text"
              value={formLogin}
              onChange={(e) => setFormLogin(e.target.value)}
              placeholder="Введите логин"
              required
            />
          </Form.Group>

          <Form.Group controlId="password" className="mb-3">
            <Form.Label>Пароль</Form.Label>
            <Form.Control
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Введите пароль"
              required
            />
          </Form.Group>

          <Button
            variant="primary"
            type="submit"
            disabled={loading}
            className="w-100"
          >
            {loading ? (
              <>
                <Spinner
                  as="span"
                  animation="border"
                  size="sm"
                  role="status"
                  aria-hidden="true"
                />{" "}
                Входим...
              </>
            ) : (
              "Войти"
            )}
          </Button>
        </Form>
      </Card>
    </Container>
  );
};

export default LoginPage;
