import { useEffect, useRef, useState } from "react";
import { Alert, Button, Card, Col, Container, Form, Row, Spinner } from "react-bootstrap";
import { axiosGetTables, axiosExportTables, axiosImportTables, type DbExportFile } from "../../requests/dbAdminRequests";

const ALL_TABLES = "all";

const DbAdminPage: React.FC = () => {
    const [tables, setTables] = useState<string[]>([]);
    const [selectedTable, setSelectedTable] = useState<string>(ALL_TABLES);
    const [exportFormat, setExportFormat] = useState<string>("json");

    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState<string | null>(null);
    const [exporting, setExporting] = useState<boolean>(false);
    const [importing, setImporting] = useState<boolean>(false);
    const [importFile, setImportFile] = useState<File | null>(null);
    const fileInputRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        const fetchTables = async () => {
            try {
                const data = await axiosGetTables();
                setTables(data || []);
            } catch {
                setError("Не удалось загрузить список таблиц");
            }
        };
        fetchTables();
    }, []);

    const handleExport = async () => {
        setError(null);
        setSuccess(null);
        setExporting(true);
        try {
            await axiosExportTables(selectedTable, exportFormat);
        } catch (err) {
            console.error(err);
            setError("Ошибка при экспорте данных");
        } finally {
            setExporting(false);
        }
    };

    const handleImport = async () => {
        setError(null);
        setSuccess(null);

        if (!importFile) {
            setError("Сначала выберите JSON-файл для импорта");
            return;
        }

        let data: DbExportFile;
        try {
            data = JSON.parse(await importFile.text());
        } catch {
            setError("Выбранный файл не является корректным JSON");
            return;
        }

        if (!data.tables || typeof data.tables !== "object" || Array.isArray(data.tables)) {
            setError("В файле отсутствует блок \"tables\" — выберите файл, полученный через экспорт");
            return;
        }

        const tableNames = Object.keys(data.tables);
        if (tableNames.length === 0) {
            setError("Файл не содержит ни одной таблицы");
            return;
        }

        const confirmed = window.confirm(
            `Импорт заменит данные в таблицах: ${tableNames.join(", ")}.\n\nТекущие данные этих таблиц будут удалены. Продолжить?`
        );
        if (!confirmed) return;

        setImporting(true);
        try {
            await axiosImportTables({ tables: data.tables });
            setSuccess(`Импорт завершён: ${tableNames.join(", ")}`);
            setImportFile(null);
            if (fileInputRef.current) fileInputRef.current.value = "";
        } catch (err: any) {
            console.error(err);
            const message = err?.response?.data?.message;
            setError(message ? `Ошибка при импорте: ${message}` : "Ошибка при импорте данных");
        } finally {
            setImporting(false);
        }
    };

    return (
        <Container className="py-4">
            <h3 className="mb-4">Экспорт и импорт данных</h3>

            {error && <Alert variant="danger" onClose={() => setError(null)} dismissible>{error}</Alert>}
            {success && <Alert variant="success" onClose={() => setSuccess(null)} dismissible>{success}</Alert>}

            <Row className="gy-4">
                <Col xs={12} md={6}>
                    <Card className="h-100 dynamic-bg-sidebar text">
                        <Card.Body>
                            <Card.Title>Экспорт</Card.Title>
                            <Card.Text className="small text-muted">
                                Скачать выбранную таблицу или всю базу данных в JSON-файл
                                или как SQL-скрипт для восстановления через psql.
                            </Card.Text>
                            <Form.Group className="mb-3">
                                <Form.Label className="fw-semibold">Таблица:</Form.Label>
                                <Form.Select
                                    value={selectedTable}
                                    onChange={(e) => setSelectedTable(e.target.value)}
                                >
                                    <option value={ALL_TABLES}>-- Все таблицы --</option>
                                    {tables.map((t) => (
                                        <option key={t} value={t}>{t}</option>
                                    ))}
                                </Form.Select>
                            </Form.Group>
                            <Form.Group className="mb-3">
                                <Form.Label className="fw-semibold">Формат:</Form.Label>
                                <Form.Select
                                    value={exportFormat}
                                    onChange={(e) => setExportFormat(e.target.value)}
                                >
                                    <option value="json">JSON</option>
                                    <option value="sql">SQL</option>
                                </Form.Select>
                            </Form.Group>
                            <Button variant="primary" className="w-100" onClick={handleExport} disabled={exporting}>
                                {exporting ? (
                                    <>
                                        <Spinner size="sm" className="me-2" />
                                        Экспорт...
                                    </>
                                ) : (
                                    "Скачать экспорт"
                                )}
                            </Button>
                        </Card.Body>
                    </Card>
                </Col>

                <Col xs={12} md={6}>
                    <Card className="h-100 dynamic-bg-sidebar text">
                        <Card.Body>
                            <Card.Title>Импорт</Card.Title>
                            <Card.Text className="small text-muted">
                                Восстановить данные из JSON-файла, полученного через экспорт.
                                Данные таблиц из файла будут полностью заменены.
                            </Card.Text>
                            <Form.Group className="mb-3">
                                <Form.Label className="fw-semibold">JSON-файл:</Form.Label>
                                <Form.Control
                                    type="file"
                                    accept=".json,application/json"
                                    ref={fileInputRef}
                                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setImportFile(e.target.files?.[0] ?? null)}
                                />
                            </Form.Group>
                            <Button variant="success" className="w-100" onClick={handleImport} disabled={importing}>
                                {importing ? (
                                    <>
                                        <Spinner size="sm" className="me-2" />
                                        Импорт...
                                    </>
                                ) : (
                                    "Импортировать данные"
                                )}
                            </Button>
                        </Card.Body>
                    </Card>
                </Col>
            </Row>
        </Container>
    );
};

export default DbAdminPage;
