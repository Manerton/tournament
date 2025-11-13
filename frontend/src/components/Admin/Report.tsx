import { useEffect, useState } from "react";
import { Alert, Container, Form, Button, Row, Col } from "react-bootstrap";
import { EventStatus, type EventResponse } from "../../types/TournamentTypes";
import { axiosGetEventByStatus } from "../../requests/tournamentRequests";
import { axiosGetLeaderBoardReport, axiosGetTeamReport } from "../../requests/reportRequests";

const ReportPage: React.FC = () => {
    const [registrationTournaments, setRegistrationTournaments] = useState<EventResponse[]>([]);
    const [closedTournaments, setClosedTournaments] = useState<EventResponse[]>([]);

    const [selectedRegistrationTournament, setSelectedRegistrationTournament] = useState<string>("");
    const [selectedClosedTournament, setSelectedClosedTournament] = useState<string>("");

    const [error, setError] = useState<string | null>(null);

    // Загружаем турниры обоих типов при монтировании
    useEffect(() => {
        const fetchTournaments = async () => {
            try {
                const [registration, closed] = await Promise.all([
                    axiosGetEventByStatus(EventStatus.Registration),
                    axiosGetEventByStatus(EventStatus.Closed),
                ]);
                setRegistrationTournaments(registration || []);
                setClosedTournaments(closed || []);
            } catch {
                setError("Не удалось загрузить список турниров");
            }
        };
        fetchTournaments();
    }, []);

    const handlerGetTeamReport = async () => {
        if (!selectedRegistrationTournament) {
            setError("Сначала выберите турнир для отчёта по командам");
            return;
        }
        try {
            await axiosGetTeamReport(selectedRegistrationTournament);
        } catch (err) {
            console.error(err);
            setError("Ошибка при загрузке отчёта по командам");
        }
    };

    const handlerGetLeaderboardReport = async () => {
        if (!selectedClosedTournament) {
            setError("Сначала выберите турнир для отчёта по результатам");
            return;
        }
        try {
            await axiosGetLeaderBoardReport(selectedClosedTournament);
        } catch (err) {
            console.error(err);
            setError("Ошибка при загрузке отчёта по результатам турнира");
        }
    };

    return (
        <Container className="py-4">
            {error && <Alert variant="danger">{error}</Alert>}

            <Row className="gy-4">
                {/* Первый селектор — для турниров с регистрацией */}
                <Col xs={12} md={6}>
                    <Form.Group>
                        <Form.Label className="fw-semibold">Отчёт по командам (турниры с регистрацией):</Form.Label>
                        <Form.Select
                            value={selectedRegistrationTournament}
                            onChange={(e) => setSelectedRegistrationTournament(e.target.value)}
                        >
                            <option value="">-- Выберите турнир --</option>
                            {registrationTournaments.map((t) => (
                                <option key={t.id} value={t.id}>
                                    {t.name} ({new Date(t.start_date).toLocaleDateString()} —{" "}
                                    {new Date(t.end_date).toLocaleDateString()})
                                </option>
                            ))}
                        </Form.Select>
                    </Form.Group>

                    <Button variant="primary" className="mt-3 w-100" onClick={handlerGetTeamReport}>
                        Скачать отчёт по командам
                    </Button>
                </Col>

                {/* Второй селектор — для закрытых турниров */}
                <Col xs={12} md={6}>
                    <Form.Group>
                        <Form.Label className="fw-semibold">Отчёт по результатам (закрытые турниры):</Form.Label>
                        <Form.Select
                            value={selectedClosedTournament}
                            onChange={(e) => setSelectedClosedTournament(e.target.value)}
                        >
                            <option value="">-- Выберите турнир --</option>
                            {closedTournaments.map((t) => (
                                <option key={t.id} value={t.id}>
                                    {t.name} ({new Date(t.start_date).toLocaleDateString()} —{" "}
                                    {new Date(t.end_date).toLocaleDateString()})
                                </option>
                            ))}
                        </Form.Select>
                    </Form.Group>

                    <Button variant="success" className="mt-3 w-100" onClick={handlerGetLeaderboardReport}>
                        Скачать отчёт по участникам
                    </Button>
                </Col>
            </Row>
        </Container>
    );
};

export default ReportPage;
