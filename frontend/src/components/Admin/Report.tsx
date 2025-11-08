import { useEffect, useState } from "react";
import { Alert, Container, Form } from "react-bootstrap"
import { EventStatus, type EventResponse } from "../../types/TournamentTypes";
import { axiosGetEventByStatus } from "../../requests/tournamentRequests";
import { axiosGetTeamReport } from "../../requests/reportRequests";

const ReportPage: React.FC = () => {


    const [tournaments, setTournaments] = useState<EventResponse[]>([]);
    const [selectedTournament, setSelectedTournament] = useState<string>("");

    const [error, setError] = useState<string | null>(null);
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

    const handlerGetReport = async () => {
        if (!selectedTournament) {
            setError("Сначала выберите турнир");
            return;
        }

        try {
            await axiosGetTeamReport(selectedTournament);
        } catch (err) {
            console.error(err);
            setError("Ошибка при загрузке отчёта");
        }
    };

    return (
        <Container className="py-4">

            {error && <Alert variant="danger">{error}</Alert>}
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


            <div>
                <button className="btn btn-primary" onClick={handlerGetReport}>Скачать отчёт</button>
            </div>

        </Container>
    )
}

export default ReportPage