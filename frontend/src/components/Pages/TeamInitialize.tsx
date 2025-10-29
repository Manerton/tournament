import { useEffect, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import { Container, Row, Col, Form, Button, Card, Table, Modal, Spinner, Alert } from "react-bootstrap";
import Select from "react-select";
import type { DistrictDTOResponse, SchoolDTOResponse } from "../../types/locationTypes";
import type { RegisterTeamRequestDTO, TeamCreateDTO } from "../../types/teamType";
import type { ParticipantCreateDTO } from "../../types/participantTypes";
import { GetDistricts, GetSchools, RegisterTeam } from "../../requests/teamRequests";
import { EventStatus, type EventResponse } from "../../types/TournamentTypes";
import { axiosGetEventById } from "../../requests/tournamentRequests";
import formatDateForInput from "../../pkg/dateFormater";


// ------------------------------------
// Валидационные функции
// ------------------------------------
const validatePhoneNumber = (phone: string): string | null => {
  if (!phone.trim()) return null; // Пустое поле - допустимо (не обязательное поле)

  // Удаляем все нецифровые символы кроме +
  const cleaned = phone.replace(/[^\d+]/g, '');

  // Проверяем российский формат: +7XXXXXXXXXX или 8XXXXXXXXXX
  const russianRegex = /^(\+7|8)\d{10}$/;

  if (!russianRegex.test(cleaned)) {
    return "Номер должен быть в формате: +7XXXXXXXXXX или 8XXXXXXXXXX";
  }

  return null;
};

const validateClass = (classNum: number): string | null => {
  if (!classNum || classNum === 0) {
    return "Класс обязателен для заполнения";
  }

  if (classNum < 1 || classNum > 11) {
    return "Класс должен быть от 1 до 11";
  }

  return null;
};


// ------------------------------------
// Компонент регистрации команды
// ------------------------------------
const TeamInitialize: React.FC = () => {

  const alertRef = useRef<HTMLDivElement | null>(null);


  const { id } = useParams<{ id: string }>();

  const [event, setEvent] = useState<EventResponse>();

  const [districts, setDistricts] = useState<DistrictDTOResponse[]>([]);
  const [schools, setSchools] = useState<SchoolDTOResponse[]>([]);
  const [selectedDistrict, setSelectedDistrict] = useState("");
  const [loadingDistricts, setLoadingDistricts] = useState(false);
  const [loadingSchools, setLoadingSchools] = useState(false);

  const [team, setTeam] = useState<TeamCreateDTO>({
    team_name: "",
    responsible_name: "",
    responsible_number: "",
    event_id: id || "",
    school_id: "",
  });

  const [participants, setParticipants] = useState<ParticipantCreateDTO[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [editIndex, setEditIndex] = useState<number | null>(null);

  const [participant, setParticipant] = useState<ParticipantCreateDTO>({
    FIO: "",
    gender: 1,
    class: 0,
    school_id: "",
    initial_score: 0,
  });

  const [submitting, setSubmitting] = useState(false);
  const [notify, setNotify] = useState<{ variant: "success" | "danger"; text: string } | null>(null);

  // Состояния для ошибок валидации
  const [validationErrors, setValidationErrors] = useState<{
    phone?: string;
    participantClass?: string;
  }>({});

  // ------------------------------------
  // Загрузка муниципалитетов
  // ------------------------------------
  useEffect(() => {
    const loadDistricts = async () => {
      setLoadingDistricts(true);
      try {
        const res: DistrictDTOResponse[] = await GetDistricts();
        setDistricts(res);
      } catch (err) {
        const message = err instanceof Error ? err.message : "Не удалось загрузить муниципалитеты";
        setNotify({ variant: "danger", text: message });
        console.error("Ошибка при загрузке муниципалитет:", err);
      } finally {
        setLoadingDistricts(false);
      }
    };
    loadDistricts();
  }, []);

  // ------------------------------------
  // Загрузка школ по выбранному муниципалитету
  // ------------------------------------
  useEffect(() => {
    const loadSchools = async () => {
      if (!selectedDistrict) {
        setSchools([]);
        setTeam((prev) => ({ ...prev, school_id: "" }));
        return;
      }

      setLoadingSchools(true);
      try {
        const res: SchoolDTOResponse[] = await GetSchools(selectedDistrict)
        setSchools(res)
      } catch (err) {
        const message = err instanceof Error ? err.message : "Не удалось загрузить школы";
        setNotify({ variant: "danger", text: message });
        console.error("Ошибка при загрузке школ:", err);
      } finally {
        setLoadingSchools(false);
      }
    };
    loadSchools();
  }, [selectedDistrict]);

  useEffect(() => {
    const loadMainEvent = async () => {
      try {
        const res: EventResponse = await axiosGetEventById(id!)
        setEvent(res)
      } catch (err) {
        const message = err instanceof Error ? err.message : "Не удалось загрузить турнир";
        setNotify({ variant: "danger", text: message });
        console.error("Ошибка при загрузке турнира:", err);
      }
    }
    loadMainEvent()
  }, [id])

  // ------------------------------------
  // Обработчики изменений с валидацией
  // ------------------------------------
  const handlePhoneChange = (phone: string) => {
    setTeam({ ...team, responsible_number: phone });
    // Валидация телефона
    const phoneError = validatePhoneNumber(phone);
    setValidationErrors(prev => ({
      ...prev,
      phone: phoneError || undefined
    }));
  };

  const handleParticipantClassChange = (classNum: number) => {
    // Ограничиваем значение от 1 до 11
    let value = classNum;
    if (value > 11) value = 11;
    if (value < 1) value = 0;

    setParticipant({ ...participant, class: value });

    // Валидация класса
    const classError = validateClass(value);
    setValidationErrors(prev => ({
      ...prev,
      participantClass: classError || undefined
    }));
  };

  // ------------------------------------
  // Работа с участниками
  // ------------------------------------
  const handleAddOrEditParticipant = () => {
    if (!participant.FIO.trim()) {
      setNotify({ variant: "danger", text: "Введите ФИО участника" });
      return;
    }

    // Проверяем валидность класса
    const classError = validateClass(participant.class);
    if (classError) {
      setValidationErrors(prev => ({ ...prev, participantClass: classError }));
      setNotify({ variant: "danger", text: classError });
      return;
    }

    const data = { ...participant, school_id: team.school_id };

    if (editIndex !== null) {
      const updated = [...participants];
      updated[editIndex] = data;
      setParticipants(updated);
      setEditIndex(null);
    } else {
      setParticipants([...participants, data]);
    }

    setParticipant({ FIO: "", gender: 1, class: 0, school_id: "", initial_score: 0 });
    setValidationErrors(prev => ({ ...prev, participantClass: undefined }));
    setShowModal(false);
  };

  const handleEdit = (index: number) => {
    setParticipant(participants[index]);
    setEditIndex(index);
    setShowModal(true);
  };

  const handleDelete = (index: number) => {
    setParticipants(participants.filter((_, i) => i !== index));
  };

  // ------------------------------------
  // Отправка команды
  // ------------------------------------
  const handleSubmit = async () => {
    // Проверяем обязательные поля
    if (!team.team_name.trim()) {
      setNotify({ variant: "danger", text: "Введите название команды" });
      return;
    }
    if (!team.school_id) {
      setNotify({ variant: "danger", text: "Выберите школу" });
      return;
    }
    if (!participants.length) {
      setNotify({ variant: "danger", text: "Добавьте хотя бы одного участника" });
      return;
    }

    // Проверяем валидность телефона (если он заполнен)
    if (team.responsible_number.trim()) {
      const phoneError = validatePhoneNumber(team.responsible_number);
      if (phoneError) {
        setValidationErrors(prev => ({ ...prev, phone: phoneError }));
        setNotify({ variant: "danger", text: phoneError });
        return;
      }
    }

    const dto: RegisterTeamRequestDTO = {
      ...team,
      participants,
    };

    try {
      setSubmitting(true);
      await RegisterTeam(dto)

      setNotify({ variant: "success", text: "Команда успешно зарегистрирована!" });
      setTeam({ ...team, team_name: "", responsible_name: "", responsible_number: "", school_id: "" });
      setParticipants([]);
      setValidationErrors({});

    } catch (err) {
      const message = err instanceof Error ? err.message : "Не удалось зарегестрировать команду";
      setNotify({ variant: "danger", text: message });
      console.error("Ошибка при загрузке регистарации команды:", err);

    } finally {
      setSubmitting(false);
    }
  };

  // ------------------------------------
  // Форматирование телефона при вводе
  // ------------------------------------
  const formatPhoneNumber = (value: string): string => {
    // Удаляем все нецифровые символы кроме +
    const numbers = value.replace(/[^\d+]/g, '');

    if (numbers.startsWith('+7')) {
      return `+7${numbers.slice(2, 12)}`;
    } else if (numbers.startsWith('8')) {
      return `8${numbers.slice(1, 11)}`;
    } else if (numbers.startsWith('7')) {
      return `+7${numbers.slice(1, 11)}`;
    }

    return numbers;
  };

  useEffect(() => {
    if (notify && alertRef.current) {
      alertRef.current.scrollIntoView({ behavior: "smooth", block: "center" });
    }
  }, [notify]);

  if (event?.status != EventStatus.Registration) {
    return (
      <Container>
        <Card className="mb-4 shadow-sm">
          <h2 >{event?.name}</h2>
          <h5>Регистрация завершена</h5>
        </Card>
      </Container>
    )
  }



  // ------------------------------------
  // JSX
  // ------------------------------------
  return (
    <Container >
      <div
        className="d-flex justify-content-center align-items-end w-100 m-0 p-0 position-relative"
        style={{ zIndex: 0 }}
      >
        <img
          src="/RST logo.png"
          alt="Логотип"
          className="ms-3"
          style={{ height: "50px", objectFit: "contain", transform: "translateY(2px)" }}
        />
        <img
          src="/battleroyale cat+.png"
          alt="Логотип"
          className="me-2"
          style={{ height: "120px", objectFit: "contain", transform: "translateY(2px)" }}
        />
      </div>



      <Card className="mb-4 shadow-sm liquid-glass-bg card-interactive">
        <Card.Body className="p-0">
          <Row>
            <Col xs={2}>

            </Col>
            <Col>
              <h6>ГАОУ АО ДО «Региональный школьный технопарк»</h6>
            </Col>
            <Col xs={2}>
            </Col>
          </Row>
          <Row>
            <Col>
              {/* Основное название мероприятия */}
              {event?.name && (
                <h3 className="mb-0">
                  {event.name.includes('«') || event.name.includes('"') ? (
                    <>
                      {event.name.split(/["«»]/)[0].trim()}
                      <br />
                      «{event.name.split(/["«»]/)[1].trim()}»
                    </>
                  ) : (
                    event.name
                  )}
                </h3>

              )}
              {event?.start_date && (
                <h6 className="mt-3 mb-0">
                  {formatDateForInput(event.start_date)}
                </h6>
              )}

            </Col>
          </Row>



        </Card.Body>

      </Card>

      {notify && (
        <div ref={alertRef}>
          <Alert variant={notify.variant}>{notify.text}</Alert>
        </div>
      )}

      <Card className="mb-4 shadow-sm liquid-glass-bg card-interactive">
        <h2 className="text-center mb-4">Регистрация команды</h2>
        <Card.Body>
          <h5 className="mb-3">Информация о команде</h5>
          <Form>
            <Form.Group className="mb-3">
              <Form.Label>Название команды *</Form.Label>
              <Form.Control
                value={team.team_name}
                onChange={(e) => setTeam({ ...team, team_name: e.target.value })}
                placeholder="Например: Лидеры будущего"
                isInvalid={!team.team_name.trim()}
              />
              <Form.Control.Feedback type="invalid">
                Название команды обязательно
              </Form.Control.Feedback>
            </Form.Group>

            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Ф.И.О. педагога, сопровождающего команду на турнире</Form.Label>
                  <Form.Control
                    value={team.responsible_name}
                    onChange={(e) => setTeam({ ...team, responsible_name: e.target.value })}
                  />
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Номер сотового телефона педагога, сопровождающего команду на турнире</Form.Label>
                  <Form.Control
                    value={team.responsible_number}
                    onChange={(e) => {
                      const formatted = formatPhoneNumber(e.target.value);
                      handlePhoneChange(formatted);
                    }}
                    placeholder="+7XXXXXXXXXX или 8XXXXXXXXXX"
                    isInvalid={!!validationErrors.phone}
                  />
                  <Form.Control.Feedback type="invalid">
                    {validationErrors.phone}
                  </Form.Control.Feedback>
                  <Form.Text className="text-muted">
                    Формат: +7XXXXXXXXXX или 8XXXXXXXXXX
                  </Form.Text>
                </Form.Group>
              </Col>
            </Row>

            <Form.Group className="mb-3">
              <Form.Label>Муниципальное образование *</Form.Label>
              <Select
                isDisabled={loadingDistricts}
                isLoading={loadingDistricts}
                placeholder="Выберите муниципалитет..."
                options={districts.map((d) => ({ value: d.id, label: d.name }))}
                value={districts
                  .filter((d) => d.id === selectedDistrict)
                  .map((d) => ({ value: d.id, label: d.name }))[0] || null}
                onChange={(opt) => {
                  setSelectedDistrict(opt ? opt.value : "");
                  setTeam((prev) => ({ ...prev, school_id: "" }));
                }}
                styles={customSelectStyles}
              />
            </Form.Group>

            <Form.Group className="mb-3">
              <Form.Label>Школа *</Form.Label>
              <Select
                isDisabled={!selectedDistrict || loadingSchools}
                isLoading={loadingSchools}
                placeholder={
                  loadingSchools
                    ? "Загрузка школ..."
                    : !selectedDistrict
                      ? "Сначала выберите муниципальное образование "
                      : "Выберите школу..."
                }
                options={schools.map((s) => ({ value: s.id, label: s.name }))}
                value={schools
                  .filter((s) => s.id === team.school_id)
                  .map((s) => ({ value: s.id, label: s.name }))[0] || null}
                onChange={(opt) => setTeam({ ...team, school_id: opt ? opt.value : "" })}
                styles={customSelectStyles}
              />
            </Form.Group>
          </Form>


          <div className="d-flex justify-content-between align-items-center mb-3">
            <h5>Участники</h5>
            <Button variant="outline-primary" onClick={() => setShowModal(true)} disabled={!team.school_id}>
              + Добавить участника
            </Button>
          </div>

          <Table striped bordered hover responsive>
            <thead>
              <tr>
                <th>Ф.И.О.</th>
                <th>Пол</th>
                <th>Класс</th>
                <th>Полный балл на школьном этапе ВСОШ по математике в 2025-2026 учебном году</th>
              </tr>
            </thead>
            <tbody>
              {participants.length > 0 ? (
                participants.map((p, i) => (
                  <tr key={i}>
                    <td>{p.FIO}</td>
                    <td>{p.gender === 1 ? "Мужской" : "Женский"}</td>
                    <td>{p.class}</td>
                    <td>{p.initial_score}</td>
                    <td>
                      <Button size="sm" variant="outline-primary" className="me-2" onClick={() => handleEdit(i)}>
                        Редактировать
                      </Button>
                      <Button size="sm" variant="outline-danger" onClick={() => handleDelete(i)}>
                        Удалить
                      </Button>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={4} className="text-center">Нет участников</td>
                </tr>
              )}
            </tbody>
          </Table>




          <div className="justify-content-center">
            <Button
              size="lg"
              variant="outline-primary"
              onClick={handleSubmit}
              disabled={submitting || !team.school_id || !participants.length}
            >
              {submitting ? <Spinner animation="border" size="sm" /> : "Отправить заявку"}
            </Button>
          </div>

          <div className="text-justify m-3 p-3  rounded" style={{ lineHeight: "1.6", fontSize: "1rem" }}>
            <p className="mb-2">
              <strong>Уважаемый педагог,</strong> регистрирующий команду! Если у вас возникли вопросы по заполнению данной формы,
              пожалуйста, обращайтесь к нам по номеру <strong>+7 988 069 1743 — Портнова Юлия Сергеевна </strong>
              или в мобильном приложении <strong>MAX</strong>.
            </p>
            <p className="mb-0">
              <em>Благодарим вас за внимание!</em>
            </p>
          </div>


        </Card.Body>
      </Card>


      {/* Модальное окно участника */}
      <Modal
        show={showModal}
        onHide={() => setShowModal(false)}
        centered
        contentClassName="dynamic-bg-modal"
        dialogClassName="glass-dialog"
      >
        <Modal.Header closeButton>
          <Modal.Title>
            {editIndex !== null ? "Редактировать участника" : "Добавить участника"}
          </Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Form>
            <Row>
              <Col md={4}>
                <Form.Group className="mb-3">
                  <Form.Label>Фамилия *</Form.Label>
                  <Form.Control
                    value={participant.FIO.split(" ")[0] || ""}
                    onChange={(e) => {
                      const parts = participant.FIO.split(" ");
                      parts[0] = e.target.value;
                      setParticipant({ ...participant, FIO: parts.join(" ").trim() });
                    }}
                    placeholder="Иванов"
                  />
                </Form.Group>
              </Col>
              <Col md={4}>
                <Form.Group className="mb-3">
                  <Form.Label>Имя *</Form.Label>
                  <Form.Control
                    value={participant.FIO.split(" ")[1] || ""}
                    onChange={(e) => {
                      const parts = participant.FIO.split(" ");
                      parts[1] = e.target.value;
                      setParticipant({ ...participant, FIO: parts.join(" ").trim() });
                    }}
                    placeholder="Иван"
                  />
                </Form.Group>
              </Col>
              <Col md={4}>
                <Form.Group className="mb-3">
                  <Form.Label>Отчество</Form.Label>
                  <Form.Control
                    value={participant.FIO.split(" ")[2] || ""}
                    onChange={(e) => {
                      const parts = participant.FIO.split(" ");
                      parts[2] = e.target.value;
                      setParticipant({ ...participant, FIO: parts.join(" ").trim() });
                    }}
                    placeholder="Иванович"
                  />
                </Form.Group>
              </Col>
            </Row>

            <Form.Group className="mb-3">
              <Form.Label>Пол *</Form.Label>
              <Form.Select
                value={participant.gender}
                onChange={(e) =>
                  setParticipant({ ...participant, gender: parseInt(e.target.value) })
                }
              >
                <option value={2}>Женский</option>
                <option value={1}>Мужской</option>
              </Form.Select>
            </Form.Group>

            <Form.Group className="mb-3">
              <Form.Label>Класс *</Form.Label>
              <Form.Control
                type="text"
                inputMode="numeric"   // открывает цифровую клавиатуру на телефоне
                pattern="[0-9]*"      // подсказывает, что нужны цифры
                value={participant.class || ""}
                onChange={(e) => {
                  const raw = e.target.value.replace(/\D/g, ""); // оставляем только цифры
                  const num = parseInt(raw, 10);
                  handleParticipantClassChange(isNaN(num) ? 0 : num);
                }}
                isInvalid={!!validationErrors.participantClass}
              />
              <Form.Control.Feedback type="invalid">
                {validationErrors.participantClass}
              </Form.Control.Feedback>
            </Form.Group>

            <Form.Group className="mb-3">
              <Form.Label>Полный балл на школьном этапе ВСОШ по математике в 2025-2026 учебном году *</Form.Label>
              <Form.Control
                type="number"
                inputMode="numeric"
                value={participant.initial_score || 0}
                onChange={(e) => {
                  const val = parseInt(e.target.value, 10);
                  const safeVal = isNaN(val) ? 0 : Math.min(Math.max(val, 0), 100); // ограничиваем 0–100
                  setParticipant({ ...participant, initial_score: safeVal });
                }}
                min={0}
                max={100}
              />
              <Form.Text className="text-muted">
                Балл от 0 до 100
              </Form.Text>
            </Form.Group>


          </Form>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="outline-secondary" onClick={() => setShowModal(false)}>
            Отмена
          </Button>
          <Button
            variant="outline-primary"
            onClick={handleAddOrEditParticipant}
            disabled={!!validationErrors.participantClass}
          >
            {editIndex !== null ? "Сохранить" : "Добавить"}
          </Button>
        </Modal.Footer>
      </Modal>

    </Container>
  );
}

export default TeamInitialize

const customSelectStyles = {
  control: (base: any) => ({
    ...base,
    borderRadius: "8px",
    backgroundColor: "var(--select-bg)",
    color: "var(--select-text)",
    borderColor: "var(--select-border)",
    boxShadow: "none",
    "&:hover": {
      borderColor: "var(--select-hover)",
    },
  }),
  menu: (base: any) => ({
    ...base,
    backgroundColor: "var(--select-bg)",
    color: "var(--select-text)",
    zIndex: 9999,
  }),
  singleValue: (base: any) => ({
    ...base,
    color: "var(--select-text)",
  }),
  input: (base: any) => ({
    ...base,
    color: "var(--select-text)",
  }),
  placeholder: (base: any) => ({
    ...base,
    color: "var(--select-text)",
    opacity: 0.7,
  }),
  option: (base: any, state: any) => ({
    ...base,
    backgroundColor: state.isFocused
      ? "var(--select-focus)"
      : "var(--select-bg)",
    color: "var(--select-text)",
    cursor: "pointer",
  }),
};
