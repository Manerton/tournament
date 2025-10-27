import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { EventStatus, type EventResponse } from "../../types/TournamentTypes";
import { Alert, Button, Card, Container, Spinner } from "react-bootstrap";
import { axiosGetTours, axiosStartTournament } from "../../requests/tournamentRequests";

const ToursPage: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const [tours, setTours] = useState<EventResponse[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [starting, setStarting] = useState(false);

    const navigate = useNavigate();

    const loadTours = async () => {
        try {
            setLoading(true);
            const res = await axiosGetTours(id!);
            setTours(res);
            setError(null);
        } catch (err) {
            console.error(err);
            setError("Ошибка при загрузке туров");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadTours();
    }, [id]);

    const handleStartTournament = async () => {
        if (!id) return;
        try {
            setStarting(true);
            await axiosStartTournament(id); // вызываем backend
            await loadTours();              // обновляем список
        } catch (err) {
            console.error(err);
            setError("Не удалось начать турнир");
        } finally {
            setStarting(false);
        }
    };

    if (loading) {
        return (
            <Container className="py-5 text-center">
                <Spinner animation="border" variant="primary" />
            </Container>
        );
    }

    if (error) {
        return (
            <Container className="py-5 text-center">
                <Alert variant="danger">{error}</Alert>
            </Container>
        );
    }

    return (
        <Container className="py-4">
            <h2 className="mb-4 fw-bold text-center display-6">Список туров</h2>

            {tours.length === 0 ? (
                <div className="d-flex flex-column align-items-center justify-content-center py-5">
                    <Button
                        variant="primary"
                        size="lg"
                        disabled={starting}
                        className="px-5 py-3 rounded-4 fw-semibold shadow-sm"
                        onClick={handleStartTournament}
                    >
                        {starting ? (
                            <>
                                <Spinner
                                    as="span"
                                    animation="border"
                                    size="sm"
                                    role="status"
                                    aria-hidden="true"
                                    className="me-2"
                                />
                                Запуск...
                            </>
                        ) : (
                            <>🚀 Начать турнир</>
                        )}
                    </Button>
                </div>
            ) : (
                <div className="d-flex flex-column gap-4">
                    {tours.map((t) => (
                        <Card
                            key={t.id}
                            className="rounded-4 shadow-sm liquid-glass-bg card-interactive"
                            onClick={() => navigate(`/tours/${t.id}/games`)}
                            style={{ cursor: "pointer" }}
                        >
                            <Card.Body className="py-4">
                                <Card.Title className="fw-bold mb-3 fs-3 text-body">
                                    {t.name}
                                </Card.Title>
                                <Card.Text className="fs-5 text-body">
                                    Статус:{" "}
                                    <span
                                        className={
                                            t.status === EventStatus.Active
                                                ? "text-success"
                                                : "text-body-secondary"
                                        }
                                    >
                                        {t.status === EventStatus.Active ? "Активен" : "Завершён"}
                                    </span>
                                </Card.Text>
                            </Card.Body>
                        </Card>
                    ))}
                </div>
            )}
        </Container>
    );
};

export default ToursPage;
