import React, { useEffect, useState } from "react";
import { Form, Table, Spinner, Alert, Container } from "react-bootstrap";
import type { EventResponse } from "../../types/TournamentTypes";
import type { TeamResponceDTO } from "../../types/teamType";
import { axiosGetTournamentActive } from "../../requests/tournamentRequests";
import { GetTeamsByEvent } from "../../requests/teamRequests";


const TeamApplicationsPage: React.FC = () => {
  const [tournaments, setTournaments] = useState<EventResponse[]>([]);
  const [selectedTournament, setSelectedTournament] = useState<string>("");
  const [teams, setTeams] = useState<TeamResponceDTO[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Загружаем список турниров при монтировании
  useEffect(() => {
    const fetchTournaments = async () => {
      try {
        const res = await axiosGetTournamentActive();
        setTournaments(res.data.data || []);
      } catch (err) {
        setError("Не удалось загрузить список турниров");
      }
    };
    fetchTournaments();
  }, []);

  // Когда выбран турнир — подгружаем команды
  useEffect(() => {
    if (!selectedTournament) return;

    const fetchTeams = async () => {
      setLoading(true);
      setError(null);
      try {
        const data = await GetTeamsByEvent(selectedTournament);
        setTeams(data);
      } catch (err: any) {
        setError(err.message || "Ошибка при загрузке команд");
      } finally {
        setLoading(false);
      }
    };

    fetchTeams();
  }, [selectedTournament]);

  return (
    <Container className="py-4">
      <h2 className="mb-4 text-center fw-bold">Список команд по турнирам</h2>

      {/* Селектор турниров */}
      <Form.Group className="mb-4">
        <Form.Label className="fw-semibold">Выберите турнир:</Form.Label>
        <Form.Select
          value={selectedTournament}
          onChange={(e) => setSelectedTournament(e.target.value)}
        >
          <option value="">-- Выберите турнир --</option>
          {tournaments.map((t) => (
            <option key={t.id} value={t.id}>
              {t.name} ({new Date(t.start_date).toLocaleDateString()} —{" "}
              {new Date(t.end_date).toLocaleDateString()})
            </option>
          ))}
        </Form.Select>
      </Form.Group>

      {/* Ошибки и загрузка */}
      {loading && <Spinner animation="border" />}
      {error && <Alert variant="danger">{error}</Alert>}

      {/* Таблица команд */}
      {teams.length > 0 && !loading && (
        <Table striped bordered hover responsive>
          <thead>
            <tr>
              <th>#</th>
              <th>Название команды</th>
              <th>Ответственный</th>
              <th>Телефон</th>
              <th>Статус</th>
            </tr>
          </thead>
          <tbody>
            {teams.map((team, idx) => (
              <tr key={idx}>
                <td>{idx + 1}</td>
                <td>{team.team_name}</td>
                <td>{team.responsible_name}</td>
                <td>{team.responsible_number}</td>
                <td>
                  {team.status === 2
                    ? "Принято"
                    : team.status === 1
                    ? "На рассмотрении"
                    : "Неизвестно"}
                </td>
              </tr>
            ))}
          </tbody>
        </Table>
      )}

      {!loading && selectedTournament && teams.length === 0 && (
        <Alert variant="info">Нет зарегистрированных команд.</Alert>
      )}
    </Container>
  );
};

export default TeamApplicationsPage;
