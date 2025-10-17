import React, { useEffect, useState } from "react";
import { Form, Table, Spinner, Alert, Container, Button } from "react-bootstrap";
import type { EventResponse } from "../../types/TournamentTypes";
import type { TeamResponceDTO } from "../../types/teamType";
import { axiosGetTournamentActive } from "../../requests/tournamentRequests";
import { GetTeamsByEvent } from "../../requests/teamRequests";
import { ChangeTeamStatus } from "../../requests/teamRequests"; // <-- импортируем функцию

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

  // 👉 Обработчик изменения статуса
  const handleChangeStatus = async (teamId: string) => {
    try {
      await ChangeTeamStatus(teamId);

      // После успешного запроса — локально обновляем статус
      setTeams((prevTeams) =>
        prevTeams.map((team) =>
          team.id === teamId
            ? {
                ...team,
                status: team.status === 1 ? 2 : 1, // Переключаем между "На рассмотрении" и "Принято"
              }
            : team
        )
      );
    } catch (err: any) {
      alert(err.message || "Ошибка при изменении статуса");
    }
  };

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
              <th>Действие</th>
            </tr>
          </thead>
          <tbody>
            {teams.map((team, idx) => (
              <tr key={team.id}>
                <td>{idx + 1}</td>
                <td>{team.name}</td>
                <td>{team.responsible_name}</td>
                <td>{team.responsible_phone}</td>
                <td>
                  {team.status === 2
                    ? "Принято"
                    : team.status === 1
                    ? "На рассмотрении"
                    : "Неизвестно"}
                </td>
                <td className="text-center">
                  <Button
                    variant={team.status === 2 ? "warning" : "success"}
                    size="sm"
                    onClick={() => handleChangeStatus(team.id)}
                  >
                    {team.status === 2 ? "Отклонить" : "Принять"}
                  </Button>
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
