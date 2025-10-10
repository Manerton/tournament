import React, { useState, useEffect } from "react";
import { useParams } from "react-router-dom";
import { Container, Row, Col, Form, Button, Card, Table, Modal } from "react-bootstrap";

// ----------------------------
// Типы данных
// ----------------------------
interface DistrictDTOResponse {
  ID: string;
  Name: string;
  RegionID: string;
}

interface SchoolDTOResponse {
  ID: string;
  Name: string;
  DistrictID: string;
}

interface Participant {
  id: string;
  fio: string;
  gender: number;
  class: number;
  school_id: string;
}

interface ParticipantCreateDTO {
  FIO: string;
  Gender: number;
  Class: number;
  SchoolID: string;
}

interface Team {
  team_name: string;
  responsible_name: string;
  responsible_number: string;
  school_id: string;
  event_id: string;
}

interface RegisterTeamRequestDTO {
  TeamName: string;
  ResponsibleName: string;
  ResponsibleNumber: string;
  EventID: string;
  SchoolID: string;
  Participants: ParticipantCreateDTO[];
}

interface ParticipantName {
  lastName: string;
  firstName: string;
  middleName: string;
}

// ----------------------------
// Компонент страницы регистрации
// ----------------------------
export default function TeamInitialize() {
  const { id } = useParams<{ id: string }>();

  const [team, setTeam] = useState<Team>({
    team_name: "",
    responsible_name: "",
    responsible_number: "",
    school_id: "",
    event_id: id || "",
  });

  const [municipalities, setMunicipalities] = useState<DistrictDTOResponse[]>([]);
  const [selectedMunicipality, setSelectedMunicipality] = useState<string>("");
  const [schools, setSchools] = useState<SchoolDTOResponse[]>([]);
  const [loadingSchools, setLoadingSchools] = useState<boolean>(false);

  const [participants, setParticipants] = useState<Participant[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [currentParticipantName, setCurrentParticipantName] = useState<ParticipantName>({
    lastName: "",
    firstName: "",
    middleName: "",
  });
  const [currentParticipantDetails, setCurrentParticipantDetails] = useState({
    gender: 0,
    class: 0,
  });
  const [editIndex, setEditIndex] = useState<number | null>(null);

  // ----------------------------
  // Загрузка муниципалитетов
  // ----------------------------
  useEffect(() => {
    const loadMunicipalities = async () => {
      try {
        const res = await fetch("http://localhost:8080/api/locations/districts");
        if (!res.ok) throw new Error("Ошибка при загрузке муниципалитетов");

        const apiResponse = await res.json() as { status: string; data: DistrictDTOResponse[] };
        if (!apiResponse.data) throw new Error("Пустой ответ");

        setMunicipalities(apiResponse.data);
      } catch (error) {
        console.error(error);
        setMunicipalities([]);
      }
    };
    loadMunicipalities();
  }, []);

  // ----------------------------
  // Загрузка школ по муниципалитету
  // ----------------------------
  useEffect(() => {
    const loadSchools = async () => {
      if (!selectedMunicipality) {
        setSchools([]);
        setTeam(prev => ({ ...prev, school_id: "" }));
        return;
      }

      setLoadingSchools(true);
      try {
        const res = await fetch(`http://localhost:8080/api/locations/districts/${selectedMunicipality}/schools`);
        if (!res.ok) throw new Error("Ошибка при загрузке школ");

        const apiResponse = await res.json() as { status: string; data: SchoolDTOResponse[] };
        if (!apiResponse.data) throw new Error("Пустой ответ");

        setSchools(apiResponse.data);
        setTeam(prev => ({ ...prev, school_id: "" }));
      } catch (error) {
        console.error(error);
        setSchools([]);
      } finally {
        setLoadingSchools(false);
      }
    };
    loadSchools();
  }, [selectedMunicipality]);

  // ----------------------------
  // Помощники для ФИО
  // ----------------------------
  const combineFIO = (name: ParticipantName): string => `${name.lastName} ${name.firstName} ${name.middleName}`.trim();
  const splitFIO = (fio: string): ParticipantName => {
    const parts = fio.split(" ");
    return { lastName: parts[0] || "", firstName: parts[1] || "", middleName: parts[2] || "" };
  };

  const handleMunicipalityChange = (municipalityId: string) => setSelectedMunicipality(municipalityId);

  // ----------------------------
  // Работа с участниками
  // ----------------------------
  const handleAddParticipant = () => {
    if (!team.school_id) return alert("Сначала выберите школу команды!");
    if (!currentParticipantName.lastName.trim() || !currentParticipantName.firstName.trim()) return alert("Заполните фамилию и имя участника!");

    const participantId = editIndex !== null ? participants[editIndex].id : crypto.randomUUID();

    const participant: Participant = {
      id: participantId,
      fio: combineFIO(currentParticipantName),
      gender: currentParticipantDetails.gender,
      class: currentParticipantDetails.class,
      school_id: team.school_id,
    };

    if (editIndex !== null) {
      const updated = [...participants];
      updated[editIndex] = participant;
      setParticipants(updated);
      setEditIndex(null);
    } else {
      setParticipants([...participants, participant]);
    }

    setCurrentParticipantName({ lastName: "", firstName: "", middleName: "" });
    setCurrentParticipantDetails({ gender: 0, class: 0 });
    setShowModal(false);
  };

  const handleEditParticipant = (index: number) => {
    const participant = participants[index];
    setCurrentParticipantName(splitFIO(participant.fio));
    setCurrentParticipantDetails({ gender: participant.gender, class: participant.class });
    setEditIndex(index);
    setShowModal(true);
  };

  const handleDeleteParticipant = (index: number) => setParticipants(participants.filter((_, i) => i !== index));

  // ----------------------------
  // Отправка команды на сервер
  // ----------------------------
  const handleSubmit = async () => {
    if (!team.school_id) return alert("Выберите школу!");
    if (!participants.length) return alert("Добавьте хотя бы одного участника!");
    if (!team.event_id) return alert("Событие не указано!");

    const dto: RegisterTeamRequestDTO = {
      TeamName: team.team_name,
      ResponsibleName: team.responsible_name,
      ResponsibleNumber: team.responsible_number,
      EventID: team.event_id,
      SchoolID: team.school_id,
      Participants: participants.map(p => ({
        FIO: p.fio,
        Gender: p.gender,
        Class: p.class,
        SchoolID: p.school_id,
      })),
    };

    try {
      const res = await fetch("http://localhost:8080/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(dto),
      });

      const apiResponse = await res.json() as { status: string; data?: any; message?: string };

      if (res.ok && apiResponse.status === "OK") {
        alert("Заявка успешно отправлена!");
        setTeam({ team_name: "", responsible_name: "", responsible_number: "", school_id: "", event_id: team.event_id });
        setParticipants([]);
      } else {
        throw new Error(apiResponse.message || "Ошибка при отправке заявки");
      }
    } catch (error) {
      console.error(error);
      alert("Произошла ошибка при отправке заявки");
    }
  };

  // ----------------------------
  // JSX
  // ----------------------------
  return (
    <Container className="my-4">
      <h2 className="text-center mb-4">Регистрация команды</h2>

      {/* Данные команды */}
      <Card className="mb-4 shadow-sm">
        <Card.Body>
          <h5 className="mb-3">Информация о команде</h5>
          <Form>
            <Row>
              <Col>
                <Form.Group className="mb-3">
                  <Form.Label>Название команды</Form.Label>
                  <Form.Control type="text" value={team.team_name} onChange={e => setTeam({ ...team, team_name: e.target.value })} />
                </Form.Group>
              </Col>
            </Row>

            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>ФИО ответственного</Form.Label>
                  <Form.Control type="text" value={team.responsible_name} onChange={e => setTeam({ ...team, responsible_name: e.target.value })} />
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Телефон ответственного</Form.Label>
                  <Form.Control type="text" value={team.responsible_number} onChange={e => setTeam({ ...team, responsible_number: e.target.value })} />
                </Form.Group>
              </Col>
            </Row>

            <Form.Group className="mb-3">
              <Form.Label>Муниципалитет *</Form.Label>
              <Form.Select value={selectedMunicipality} onChange={e => handleMunicipalityChange(e.target.value)}>
                <option value="">Выберите муниципалитет</option>
                {municipalities.map(m => <option key={m.ID} value={m.ID}>{m.Name}</option>)}
              </Form.Select>
            </Form.Group>

            <Form.Group className="mb-3">
              <Form.Label>Школа *</Form.Label>
              <Form.Select value={team.school_id} onChange={e => setTeam({ ...team, school_id: e.target.value })} disabled={!selectedMunicipality || loadingSchools}>
                <option value="">
                  {loadingSchools ? "Загрузка школ..." : !selectedMunicipality ? "Сначала выберите муниципалитет" : "Выберите школу"}
                </option>
                {schools.map(s => <option key={s.ID} value={s.ID}>{s.Name}</option>)}
              </Form.Select>
              {loadingSchools && <Form.Text className="text-muted">Загружаем список школ...</Form.Text>}
            </Form.Group>
          </Form>
        </Card.Body>
      </Card>

      {/* Участники */}
      <Card className="mb-4 shadow-sm">
        <Card.Body>
          <div className="d-flex justify-content-between align-items-center mb-3">
            <h5>Участники</h5>
            <Button onClick={() => setShowModal(true)} disabled={!team.school_id}>+ Добавить</Button>
          </div>

          <Table striped bordered hover>
            <thead>
              <tr>
                <th>ФИО</th>
                <th>Пол</th>
                <th>Класс</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {participants.map((p) => (
                <tr key={p.id}>
                  <td>{p.fio}</td>
                  <td>{p.gender === 1 ? "Мужской" : "Женский"}</td>
                  <td>{p.class}</td>
                  <td>
                    <Button size="sm" variant="outline-primary" className="me-2" onClick={() => handleEditParticipant(participants.findIndex(part => part.id === p.id))}>Редактировать</Button>
                    <Button size="sm" variant="outline-danger" onClick={() => handleDeleteParticipant(participants.findIndex(part => part.id === p.id))}>Удалить</Button>
                  </td>
                </tr>
              ))}
              {participants.length === 0 && (
                <tr>
                  <td colSpan={4} className="text-center">Участников пока нет</td>
                </tr>
              )}
            </tbody>
          </Table>
        </Card.Body>
      </Card>

      {/* Кнопка отправки */}
      <div className="text-center">
        <Button size="lg" onClick={handleSubmit} disabled={!team.school_id || !team.event_id || !participants.length}>Отправить заявку</Button>
      </div>

      {/* Модалка участника */}
      <Modal show={showModal} onHide={() => setShowModal(false)} centered>
        <Modal.Header closeButton>
          <Modal.Title>{editIndex !== null ? "Редактировать участника" : "Добавить участника"}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Form>
            <Row>
              <Col md={4}>
                <Form.Group className="mb-3">
                  <Form.Label>Фамилия *</Form.Label>
                  <Form.Control type="text" value={currentParticipantName.lastName} onChange={e => setCurrentParticipantName({ ...currentParticipantName, lastName: e.target.value })} placeholder="Иванов" />
                </Form.Group>
              </Col>
              <Col md={4}>
                <Form.Group className="mb-3">
                  <Form.Label>Имя *</Form.Label>
                  <Form.Control type="text" value={currentParticipantName.firstName} onChange={e => setCurrentParticipantName({ ...currentParticipantName, firstName: e.target.value })} placeholder="Иван" />
                </Form.Group>
              </Col>
              <Col md={4}>
                <Form.Group className="mb-3">
                  <Form.Label>Отчество</Form.Label>
                  <Form.Control type="text" value={currentParticipantName.middleName} onChange={e => setCurrentParticipantName({ ...currentParticipantName, middleName: e.target.value })} placeholder="Иванович" />
                </Form.Group>
              </Col>
            </Row>

            <Form.Group className="mb-3">
              <Form.Label>Пол</Form.Label>
              <Form.Select value={currentParticipantDetails.gender} onChange={e => setCurrentParticipantDetails({ ...currentParticipantDetails, gender: parseInt(e.target.value) })}>
                <option value={0}>Женский</option>
                <option value={1}>Мужской</option>
              </Form.Select>
            </Form.Group>

            <Form.Group className="mb-3">
              <Form.Label>Класс</Form.Label>
              <Form.Control type="number" value={currentParticipantDetails.class} onChange={e => setCurrentParticipantDetails({ ...currentParticipantDetails, class: parseInt(e.target.value) || 0 })} />
            </Form.Group>
          </Form>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowModal(false)}>Отмена</Button>
          <Button onClick={handleAddParticipant}>{editIndex !== null ? "Сохранить" : "Добавить"}</Button>
        </Modal.Footer>
      </Modal>
    </Container>
  );
}