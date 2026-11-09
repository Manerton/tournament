import { Card, Button } from "react-bootstrap";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../helpers/AuthContext";

const AccessDenied: React.FC = () => {
    const navigate = useNavigate();
    const { user } = useAuth();

    return (
        <div className="d-flex justify-content-center align-items-center py-5">
            <Card
                className="p-4 shadow-lg liquid-glass-bg text-center"
                style={{ minWidth: "320px", maxWidth: "420px" }}
            >
                <div className="display-4 mb-3">🚫</div>
                <Card.Title className="fw-bold fs-3 text-body mb-3">Доступ запрещён</Card.Title>
                <Card.Text className="text-body-secondary mb-4">
                    У вас нет прав для просмотра этой страницы.
                    {!user && " Войдите в систему с учётной записью администратора."}
                </Card.Text>
                <div className="d-flex gap-2 justify-content-center">
                    <Button variant="primary" onClick={() => navigate("/")}>
                        На главную
                    </Button>
                    {!user && (
                        <Button variant="outline-primary" onClick={() => navigate("/login")}>
                            Войти
                        </Button>
                    )}
                </div>
            </Card>
        </div>
    );
};

export default AccessDenied;
