// src/pages/admin/UsersManagement.tsx
import React, { useEffect, useState } from "react";
import { Table, Button, Form, Modal, Spinner, Alert } from "react-bootstrap";
import axios from "axios";
import { API_URL } from "../../config/routes";
import { ResponseStatus, type ApiResponse } from "../../types/responseTypes";
import { getCRFCookie } from "../../components/helpers/AuthContext";

// -------- axios instance with CSRF interceptor ----------
const API_BASE = API_URL || "http://172.16.0.196:8080";

const convertRoleToString = (role: number): string => {
    if (role === 1) {
        return "Администратор";
    } else if (role === 2) {
        return "Жюри";
    } else {
        return "Не указано";
    }
}

const api = axios.create({
    baseURL: API_BASE,
    withCredentials: true,
    headers: {
        "Content-Type": "application/json",
    },
});

// Интерцептор для добавления X-CSRF-Token если он есть
api.interceptors.request.use((config) => {
    try {
        const csrfToken = getCRFCookie();
        if (csrfToken && config.headers) {
            config.headers["X-CSRF-Token"] = csrfToken;
        }
    } catch (e) {
        // безопасно игнорируем — если getCRFCookie бросит, запрос всё равно уйдёт без токена
        // можно логировать при необходимости
        // console.warn("CSRF token read failed", e);
    }
    return config;
});

// -------- типы ----------
interface UserResponseDTO {
    id: string;
    login: string;
    role: number;
}

interface UserCreateDTO {
    login: string;
    password: string;
    role: number;
}

interface UserUpdateDTO {
    login?: string;
    password?: string;
    role?: number;
}

// -------- компонент ----------
const UsersManagement: React.FC = () => {
    const [users, setUsers] = useState<UserResponseDTO[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");
    const [showModal, setShowModal] = useState(false);
    const [editUser, setEditUser] = useState<UserResponseDTO | null>(null);
    const [form, setForm] = useState<UserCreateDTO>({ login: "", password: "", role: 0 });

    // ---- загрузка списка ----
    async function loadUsers() {
        setLoading(true);
        setError("");
        try {
            const res = await api.get<ApiResponse<UserResponseDTO[]>>("/api/users/all");
            const json = res.data; // <-- это объект { status, status_code, data, message? }

            const ok =
                json.status === ResponseStatus.Ok || // "Ok" из фронтенда
                json.status === "OK" ||               // "OK" от бэка
                json.status_code === 0;               // или по коду

            if (!ok) {
                throw new Error(json.message || "Ошибка при загрузке пользователей");
            }

            const usersList = Array.isArray(json.data) ? json.data : [];
            setUsers(usersList);
        } catch (err: any) {
            setError(err.response?.data?.message || err.message || "Ошибка сети");
        } finally {
            setLoading(false);
        }
    }

    useEffect(() => {
        loadUsers();
    }, []);

    // ---- создать пользователя ----
    async function handleCreateUser(e: React.FormEvent) {
        e.preventDefault();
        try {
            await api.post<ApiResponse<any>>("/api/users", form);
            setShowModal(false);
            setForm({ login: "", password: "", role: 0 });
            await loadUsers();
        } catch (err: any) {
            alert(err.response?.data?.message || "Ошибка при создании пользователя");
        }
    }

    // ---- обновить пользователя ----
    async function handleUpdateUser(id: string) {
        try {
            // DTO для обновления: отправляем только заданные поля
            const dto: UserUpdateDTO = {};
            if (form.login) dto.login = form.login;
            if (form.password) dto.password = form.password;
            if (form.role !== 0) dto.role = form.role;

            await api.put<ApiResponse<any>>(`/api/users/${id}`, dto);
            setShowModal(false);
            setEditUser(null);
            setForm({ login: "", password: "", role: 0 });
            await loadUsers();
        } catch (err: any) {
            alert(err.response?.data?.message || "Ошибка при обновлении пользователя");
        }
    }

    // ---- удалить пользователя ----
    async function handleDeleteUser(id: string) {
        if (!window.confirm("Удалить пользователя?")) return;
        try {
            await api.delete<ApiResponse<any>>(`/api/users/${id}`);
            await loadUsers();
        } catch (err: any) {
            alert(err.response?.data?.message || "Ошибка при удалении пользователя");
        }
    }

    // ---- открыть модалку для добавления/редактирования ----
    function openModal(user?: UserResponseDTO) {
        if (user) {
            setEditUser(user);
            setForm({
                login: user.login,
                password: "",
                role: user.role
            });
        } else {
            setEditUser(null);
            setForm({ login: "", password: "", role: 0 });
        }
        setShowModal(true);
    }

    return (
        <div className="container mt-4">
            <div className="d-flex justify-content-between align-items-center mb-3">
                <h3>Управление пользователями</h3>
                <Button onClick={() => openModal()}>Добавить пользователя</Button>
            </div>

            {error && <Alert variant="danger">{error}</Alert>}

            {loading ? (
                <div className="text-center"><Spinner animation="border" /></div>
            ) : (
                <Table striped bordered hover>
                    <thead>
                    <tr>
                        <th>ID</th>
                        <th>Логин</th>
                        <th>Роль</th>
                        <th>Действия</th>
                    </tr>
                    </thead>
                    <tbody>
                    {users.length === 0 ? (
                        <tr><td colSpan={4} className="text-center">Нет пользователей</td></tr>
                    ) : (
                        users.map((u) => (
                            <tr key={u.id}>
                                <td>{u.id}</td>
                                <td>{u.login}</td>
                                <td>{convertRoleToString(u.role)}</td>
                                <td>
                                    <Button
                                        variant="outline-primary"
                                        size="sm"
                                        className="me-2"
                                        onClick={() => openModal(u)}
                                    >
                                        Изменить
                                    </Button>
                                    <Button
                                        variant="outline-danger"
                                        size="sm"
                                        onClick={() => handleDeleteUser(u.id)}
                                    >
                                        Удалить
                                    </Button>
                                </td>
                            </tr>
                        ))
                    )}
                    </tbody>
                </Table>
            )}

            <Modal show={showModal} onHide={() => setShowModal(false)}>
                <Modal.Header closeButton>
                    <Modal.Title>{editUser ? "Изменить пользователя" : "Добавить пользователя"}</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <Form onSubmit={editUser ? (e) => { e.preventDefault(); handleUpdateUser(editUser.id); } : handleCreateUser}>
                        <Form.Group className="mb-3">
                            <Form.Label>Логин</Form.Label>
                            <Form.Control
                                type="text"
                                value={form.login}
                                onChange={(e) => setForm({ ...form, login: e.target.value })}
                                required
                            />
                        </Form.Group>
                        <Form.Group className="mb-3">
                            <Form.Label>Пароль</Form.Label>
                            <Form.Control
                                type="password"
                                value={form.password}
                                onChange={(e) => setForm({ ...form, password: e.target.value })}
                                required={!editUser}
                                placeholder={editUser ? "Оставьте пустым, если не нужно менять" : ""}
                            />
                        </Form.Group>
                        <Form.Group className="mb-3">
                            <Form.Label>Роль</Form.Label>
                            <Form.Select
                                value={form.role}
                                onChange={(e) => setForm({ ...form, role: parseInt(e.target.value) })}
                                required
                            >
                                <option value={0}>Выберите роль</option>
                                <option value={1}>Администратор</option>
                                <option value={2}>Жюри</option>
                            </Form.Select>
                            {form.role !== 0 && (
                                <Form.Text className="text-muted">
                                    Выбрана роль: {convertRoleToString(form.role)}
                                </Form.Text>
                            )}
                        </Form.Group>
                        <div className="text-end">
                            <Button variant="secondary" onClick={() => setShowModal(false)} className="me-2">
                                Отмена
                            </Button>
                            <Button variant="primary" type="submit">
                                {editUser ? "Сохранить" : "Создать"}
                            </Button>
                        </div>
                    </Form>
                </Modal.Body>
            </Modal>
        </div>
    );
};

export default UsersManagement;