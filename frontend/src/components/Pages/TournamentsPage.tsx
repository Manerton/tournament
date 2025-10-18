import React, { useEffect, useState } from "react";
import { Container, Card, Spinner, Alert } from "react-bootstrap";
import { useNavigate } from "react-router-dom";
import { axiosGetTournamentActive } from "../../requests/tournamentRequests";
import { EventStatus, type EventResponse } from "../../types/TournamentTypes";
import "../../assets/TournamentPage.css";

const TournamentsPage: React.FC = () => {
  const [tournaments, setTournaments] = useState<EventResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    const fetchTournaments = async () => {
      try {
        const res = await axiosGetTournamentActive();
        setTournaments(res.data.data || []);
      } catch (err: any) {
        setError("Не удалось загрузить список турниров");
      } finally {
        setLoading(false);
      }
    };
    fetchTournaments();
  }, []);

  if (loading)
    return (
      <Container className="py-5 text-center">
        <Spinner animation="border" />
      </Container>
    );

  if (error)
    return (
      <Container className="py-5">
        <Alert variant="danger">{error}</Alert>
      </Container>
    );

  return (
    <Container className="py-4 ">
      <h2 className="mb-4 fw-bold text-center display-6">Список турниров</h2>
      {tournaments.length === 0 ? (
        <Alert variant="info">Нет доступных турниров</Alert>
      ) : (
        <div className="d-flex flex-column gap-4">
          {tournaments.map((t) => (
          <Card
            key={t.id}
            className="tournament-card shadow-sm"
            onClick={() => navigate(`/tournament/registerTeam/${t.id}`)}
          >
            <Card.Body className="py-4">
              <Card.Title className="fw-bold mb-3 fs-1 text-body">
                {t.name}
              </Card.Title>
              <Card.Text className="fs-3 mb-2 text-body-secondary">
                {new Date(t.start_date).toLocaleDateString()} —{" "}
                {new Date(t.end_date).toLocaleDateString()}
              </Card.Text>
              <Card.Text className="fs-3 text-body">
                Статус:{" "}
                <span className={t.status === EventStatus.Active ? "text-success" : "text-body-secondary"}>
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

export default TournamentsPage;
