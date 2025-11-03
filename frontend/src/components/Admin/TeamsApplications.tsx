import React, { useEffect, useState } from "react";
import {
  Form,
  Spinner,
  Alert,
  Container,
  Button,
  Card,
  Collapse,
} from "react-bootstrap";
import { EventStatus, type EventResponse } from "../../types/TournamentTypes";
import { TeamStatus, type TeamResponseDTO } from "../../types/teamType";
import { axiosGetEventByStatus } from "../../requests/tournamentRequests";
import {
  GetTeamsByEvent,
  GetTeamMembers,
  ChangeTeamStatus,
} from "../../requests/teamRequests";
import type { ParticipantResponseDTO } from "../../types/participantTypes";

const TeamApplicationsPage: React.FC = () => {
  const [tournaments, setTournaments] = useState<EventResponse[]>([]);
  const [selectedTournament, setSelectedTournament] = useState<string>("");
  const [teams, setTeams] = useState<TeamResponseDTO[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [expandedTeam, setExpandedTeam] = useState<string | null>(null);
  const [teamMembers, setTeamMembers] = useState<
    Record<string, ParticipantResponseDTO[]>
  >({});
  const [loadingMembers, setLoadingMembers] = useState<string | null>(null);

  // Загружаем турниры при монтировании
  useEffect(() => {
    const fetchTournaments = async () => {
      try {
        const res = await axiosGetEventByStatus(EventStatus.Registration);
        setTournaments(res || []);
      } catch {
        setError("Не удалось загрузить список турниров");
      }
    };
    fetchTournaments();
  }, []);

  // Загружаем команды по выбранному турниру
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

  // Запрос на участников
  const handleToggleMembers = async (teamId: string) => {
    if (expandedTeam === teamId) {
      setExpandedTeam(null);
      return;
    }

    if (teamMembers[teamId]) {
      setExpandedTeam(teamId);
      return;
    }

    setLoadingMembers(teamId);
    try {
      const data = await GetTeamMembers(teamId);
      setTeamMembers((prev) => ({ ...prev, [teamId]: data }));
      setExpandedTeam(teamId);
    } catch (err) {
      console.error("Ошибка загрузки участников:", err);
    } finally {
      setLoadingMembers(null);
    }
  };

  // Изменение статуса
  const handleChangeStatus = async (teamId: string) => {
    try {
      await ChangeTeamStatus(teamId);
      setTeams((prev) =>
        prev.map((team) =>
          team.id === teamId
            ? {
                ...team,
                status:
                  team.status === TeamStatus.Review
                    ? TeamStatus.Accept
                    : TeamStatus.Review,
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
      <h2 className="mb-4 fw-bold text-start ">Список команд по турнирам</h2>

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

      {loading && <Spinner animation="border" />}
      {error && <Alert variant="danger">{error}</Alert>}

      {/* Список команд */}
      <div className="d-flex flex-column gap-3">
        {teams.map((team) => (
          <Card key={team.id} className="shadow-sm rounded-3 p-0 liquid-glass-bg card-interactive">
            <Card.Body>
              <div className="d-flex justify-content-between  flex-wrap">
                <div className="text-start">
                  <h5 className="fw-bold mb-1">{team.name}</h5>
                  <div className="text-muted small">
                    Ответственный: {team.responsible_name} • {" "}
                    {team.responsible_phone}
                  </div>
                  <div className="mt-1">
                    <span
                      className={`badge ${
                        team.status === TeamStatus.Accept
                          ? "bg-success"
                          : team.status === TeamStatus.Review
                          ? "bg-warning text-dark"
                          : "bg-secondary"
                      }`}
                    >
                      {team.status === TeamStatus.Accept
                        ? "Принято"
                        : team.status === TeamStatus.Review
                        ? "На рассмотрении"
                        : "Неизвестно"}
                    </span>
                  </div>
                </div>

                <div className="d-flex gap-2 mt-2 mt-md-0">
                  <Button
                    variant="outline-primary"
                    size="sm"
                    onClick={() => handleToggleMembers(team.id)}
                    disabled={loadingMembers === team.id}
                  >
                    {loadingMembers === team.id ? (
                      <Spinner animation="border" size="sm" />
                    ) : expandedTeam === team.id ? (
                      "Скрыть участников"
                    ) : (
                      "Показать участников"
                    )}
                  </Button>

                  <Button
                    variant={
                      team.status === TeamStatus.Accept
                        ? "outline-danger"
                        : "outline-success"
                    }
                    size="sm"
                    onClick={() => handleChangeStatus(team.id)}
                  >
                    {team.status === TeamStatus.Accept
                      ? "Отклонить"
                      : "Принять"}
                  </Button>
                </div>
              </div>

              <Collapse in={expandedTeam === team.id}>
                <div className="mt-3 ps-3 border-start border-3 border-primary">
                  {teamMembers[team.id] ? (
                    teamMembers[team.id].length > 0 ? (
                      <ul className="list-unstyled mb-0">
                        {teamMembers[team.id].map((m) => (
                          <li key={m.id} className="mb-1">
                            <strong>{m.FIO}</strong> — {m.class} класс — стартовый балл {m.initial_score}
                          </li>
                        ))}
                      </ul>
                    ) : (
                      <p className="text-muted">Нет участников</p>
                    )
                  ) : (
                    <p className="text-muted">Загрузка...</p>
                  )}
                </div>
              </Collapse>
            </Card.Body>
          </Card>
        ))}
      </div>

      {!loading && selectedTournament && teams.length === 0 && (
        <Alert variant="info" className="mt-4">
          Нет зарегистрированных команд.
        </Alert>
      )}
    </Container>
  );
};

export default TeamApplicationsPage;
