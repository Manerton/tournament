import React, { useEffect, useState } from "react";
import { Container, Card, Spinner, Alert } from "react-bootstrap";
import { useNavigate } from "react-router-dom";
import { axiosGetTournamentActive } from "../../requests/tournamentRequests";
import type { EventResponse } from "../../types/TournamentTypes";
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
              className="tournament-card shadow-sm text-center"
              onClick={() => navigate(`/tournament/registerTeam/${t.id}`)}
            >
              <Card.Body className="py-4">
                <Card.Title className="text fw-bold mb-3 fs-1">
                  {t.name}
                </Card.Title>
                <Card.Text className="text text-muted fs-3 mb-2">
                  {new Date(t.start_date).toLocaleDateString()} —{" "}
                  {new Date(t.end_date).toLocaleDateString()}
                </Card.Text>
                <Card.Text className="text fs-3">
                  Статус:{" "}
                  <span
                    className={
                      t.status === 1 ? "text-success" : "text-secondary"
                    }
                  >
                    {t.status === 1 ? "Активен" : "Завершён"}
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
