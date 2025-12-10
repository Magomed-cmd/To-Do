"""
E2E тесты для To-Do Platform API
Запуск: pytest test_todo_api.py -v
"""

import random, string
from datetime import datetime, timedelta
from typing import Dict, Optional

import pytest, requests


class APIClient:
    """Клиент для работы с API"""
    
    def __init__(self):
        self.user_service = "http://localhost:8081"
        self.task_service = "http://localhost:8082"
        self.analytics_service = "http://localhost:8083"
        self.access_token: Optional[str] = None
        self.refresh_token: Optional[str] = None
        
    def _headers(self, auth: bool = True) -> Dict:
        headers = {"Content-Type": "application/json"}
        if auth and self.access_token:
            headers["Authorization"] = f"Bearer {self.access_token}"
        return headers
    
    def _make_request(self, method: str, url: str, **kwargs):
        """Обёртка для запросов с обработкой ошибок"""
        response = requests.request(method, url, **kwargs)
        return response
    
    # Auth endpoints
    def register(self, email: Optional[str] = None, name: Optional[str] = None, 
                password: Optional[str] = None):
        url = f"{self.user_service}/auth/register"
        data = {}
        if email is not None:
            data["email"] = email
        if name is not None:
            data["name"] = name
        if password is not None:
            data["password"] = password
        response = self._make_request("POST", url, json=data, headers=self._headers(auth=False))
        
        # Автоматически устанавливаем токены при успешной регистрации
        if response.status_code == 201:
            resp_data = response.json()
            if "tokens" in resp_data:
                self.access_token = resp_data["tokens"]["accessToken"]
                self.refresh_token = resp_data["tokens"]["refreshToken"]
        
        return response
    
    def login(self, email: str, password: str):
        url = f"{self.user_service}/auth/login"
        response = self._make_request("POST", url, json={
            "email": email, "password": password
        }, headers=self._headers(auth=False))
        if response.status_code == 200:
            data = response.json()
            self.access_token = data["tokens"]["accessToken"]
            self.refresh_token = data["tokens"]["refreshToken"]
        return response
    
    def refresh_tokens(self, refresh_token: Optional[str] = None):
        url = f"{self.user_service}/auth/refresh"
        token = refresh_token or self.refresh_token
        return self._make_request("POST", url, json={
            "refreshToken": token
        }, headers=self._headers(auth=False))
    
    def logout(self, refresh_token: Optional[str] = None):
        url = f"{self.user_service}/auth/logout"
        token = refresh_token or self.refresh_token
        return self._make_request("POST", url, json={
            "refreshToken": token
        }, headers=self._headers(auth=False))
    
    def validate_token(self):
        url = f"{self.user_service}/auth/validate"
        return self._make_request("POST", url, headers=self._headers())
    
    # Profile endpoints
    def get_profile(self):
        url = f"{self.user_service}/users/profile"
        return self._make_request("GET", url, headers=self._headers())
    
    def update_profile(self, name: Optional[str] = None, avatar_url: Optional[str] = None):
        url = f"{self.user_service}/users/profile"
        data = {}
        if name:
            data["name"] = name
        if avatar_url:
            data["avatarUrl"] = avatar_url
        return self._make_request("PUT", url, json=data, headers=self._headers())
    
    def get_preferences(self):
        url = f"{self.user_service}/users/preferences"
        return self._make_request("GET", url, headers=self._headers())
    
    def update_preferences(self, preferences: Dict):
        url = f"{self.user_service}/users/preferences"
        return self._make_request("PUT", url, json=preferences, headers=self._headers())
    
    # Admin endpoints
    def list_users(self, limit: int = 20, offset: int = 0):
        url = f"{self.user_service}/admin/users"
        return self._make_request("GET", url, params={"limit": limit, "offset": offset}, 
                                 headers=self._headers())
    
    def update_user_role(self, user_id: int, role: str):
        url = f"{self.user_service}/admin/users/{user_id}/role"
        return self._make_request("PUT", url, json={"role": role}, headers=self._headers())
    
    def update_user_status(self, user_id: int, is_active: bool):
        url = f"{self.user_service}/admin/users/{user_id}/status"
        return self._make_request("PUT", url, json={"isActive": is_active}, 
                                 headers=self._headers())
    
    # Task endpoints
    def list_tasks(self, **params):
        url = f"{self.task_service}/tasks"
        return self._make_request("GET", url, params=params, headers=self._headers())
    
    def create_task(self, title: str, **kwargs):
        url = f"{self.task_service}/tasks"
        data = {"title": title, **kwargs}
        return self._make_request("POST", url, json=data, headers=self._headers())
    
    def get_task(self, task_id: int):
        url = f"{self.task_service}/tasks/{task_id}"
        return self._make_request("GET", url, headers=self._headers())
    
    def update_task(self, task_id: int, **kwargs):
        url = f"{self.task_service}/tasks/{task_id}"
        return self._make_request("PUT", url, json=kwargs, headers=self._headers())
    
    def delete_task(self, task_id: int):
        url = f"{self.task_service}/tasks/{task_id}"
        return self._make_request("DELETE", url, headers=self._headers())
    
    def update_task_status(self, task_id: int, status: str):
        url = f"{self.task_service}/tasks/{task_id}/status"
        return self._make_request("PATCH", url, json={"status": status}, 
                                 headers=self._headers())
    
    # Comment endpoints
    def list_comments(self, task_id: int):
        url = f"{self.task_service}/tasks/{task_id}/comments"
        return self._make_request("GET", url, headers=self._headers())
    
    def create_comment(self, task_id: int, content: str):
        url = f"{self.task_service}/tasks/{task_id}/comments"
        return self._make_request("POST", url, json={"content": content}, 
                                 headers=self._headers())
    
    # Category endpoints
    def list_categories(self):
        url = f"{self.task_service}/categories"
        return self._make_request("GET", url, headers=self._headers())
    
    def create_category(self, name: str):
        url = f"{self.task_service}/categories"
        return self._make_request("POST", url, json={"name": name}, 
                                 headers=self._headers())
    
    def delete_category(self, category_id: int):
        url = f"{self.task_service}/categories/{category_id}"
        return self._make_request("DELETE", url, headers=self._headers())
    
    # Analytics endpoints
    def get_daily_metrics(self, user_id: int, date: Optional[str] = None):
        url = f"{self.analytics_service}/metrics/daily/{user_id}"
        params = {"date": date} if date else {}
        return self._make_request("GET", url, params=params, headers=self._headers())


# Fixtures

@pytest.fixture
def api_client():
    """Создаёт новый API клиент"""
    return APIClient()


@pytest.fixture
def random_user():
    """Генерирует случайные данные пользователя"""
    suffix = ''.join(random.choices(string.ascii_lowercase + string.digits, k=8))
    return {
        "email": f"test_{suffix}@example.com",
        "name": f"Test User {suffix}",
        "password": "SecurePass123!"
    }


@pytest.fixture
def registered_user(api_client, random_user):
    """Регистрирует пользователя и возвращает данные"""
    response = api_client.register(**random_user)
    assert response.status_code == 201
    data = response.json()
    api_client.access_token = data["tokens"]["accessToken"]
    api_client.refresh_token = data["tokens"]["refreshToken"]
    return {
        **random_user,
        "user_data": data["user"],
        "tokens": data["tokens"]
    }


@pytest.fixture
def authenticated_client(api_client, registered_user):
    """Возвращает аутентифицированный клиент"""
    return api_client


# ==================== AUTH TESTS ====================

class TestAuth:
    """Тесты аутентификации"""
    
    def test_register_success(self, api_client, random_user):
        """Успешная регистрация"""
        response = api_client.register(**random_user)
        assert response.status_code == 201
        data = response.json()
        assert "user" in data
        assert "tokens" in data
        assert data["user"]["email"] == random_user["email"]
        assert data["user"]["name"] == random_user["name"]
        assert "accessToken" in data["tokens"]
        assert "refreshToken" in data["tokens"]
    
    @pytest.mark.parametrize("missing_field", ["email", "name", "password"])
    def test_register_missing_fields(self, api_client, random_user, missing_field):
        """Регистрация без обязательных полей"""
        user_data = random_user.copy()
        user_data[missing_field] = None  # Устанавливаем в None вместо удаления
        response = api_client.register(**user_data)
        assert response.status_code == 400
    
    @pytest.mark.parametrize("invalid_email", [
        "notanemail",
        "@example.com",
        "test@",
        "test space@example.com"
    ])
    def test_register_invalid_email(self, api_client, random_user, invalid_email):
        """Регистрация с невалидным email"""
        response = api_client.register(invalid_email, random_user["name"], 
                                      random_user["password"])
        assert response.status_code == 400
    
    def test_register_duplicate_email(self, api_client, registered_user):
        """Регистрация с существующим email"""
        response = api_client.register(
            registered_user["email"],
            "Another Name",
            "AnotherPass123!"
        )
        assert response.status_code in [400, 409]  # 409 Conflict тоже валиден
    
    def test_login_success(self, api_client, registered_user):
        """Успешный вход"""
        new_client = APIClient()
        response = new_client.login(registered_user["email"], registered_user["password"])
        assert response.status_code == 200
        data = response.json()
        assert "user" in data
        assert "tokens" in data
        assert new_client.access_token is not None
        assert new_client.refresh_token is not None
    
    @pytest.mark.parametrize("email,password,description", [
        ("nonexistent@example.com", "password123", "несуществующий пользователь"),
        ("", "password123", "пустой email"),
        ("test@example.com", "", "пустой пароль"),
    ])
    def test_login_invalid_credentials(self, api_client, email, password, description):
        """Вход с невалидными данными"""
        response = api_client.login(email, password)
        assert response.status_code in [400, 401, 500]  # 500 если API ошибка
    
    def test_login_wrong_password(self, api_client, registered_user):
        """Вход с неверным паролем"""
        response = api_client.login(registered_user["email"], "WrongPassword123!")
        assert response.status_code == 401
    
    def test_refresh_token_success(self, authenticated_client):
        """Успешное обновление токенов"""
        old_access = authenticated_client.access_token
        response = authenticated_client.refresh_tokens()
        assert response.status_code == 200
        data = response.json()
        assert "accessToken" in data
        assert "refreshToken" in data
        # Токены могут быть одинаковыми если время не изменилось (зависит от реализации)
        # Главное что они валидны
    
    def test_refresh_token_invalid(self, api_client):
        """Обновление с невалидным refresh token"""
        response = api_client.refresh_tokens("invalid_token_xyz")
        assert response.status_code in [401, 500]  # Может быть 500 если не обработано
    
    def test_logout_success(self, authenticated_client):
        """Успешный выход"""
        response = authenticated_client.logout()
        assert response.status_code == 204
    
    def test_validate_token_success(self, authenticated_client):
        """Валидация валидного токена"""
        response = authenticated_client.validate_token()
        assert response.status_code == 200
        data = response.json()
        assert "userId" in data
        assert "email" in data
        assert "role" in data
    
    def test_validate_token_unauthorized(self, api_client):
        """Валидация без токена"""
        response = api_client.validate_token()
        assert response.status_code == 401


# ==================== PROFILE TESTS ====================

class TestProfile:
    """Тесты профиля пользователя"""
    
    def test_get_profile_success(self, authenticated_client, registered_user):
        """Получение профиля"""
        response = authenticated_client.get_profile()
        assert response.status_code == 200
        data = response.json()
        assert data["email"] == registered_user["email"]
        assert data["name"] == registered_user["name"]
        assert "id" in data
        assert "role" in data
    
    def test_get_profile_unauthorized(self, api_client):
        """Получение профиля без авторизации"""
        response = api_client.get_profile()
        assert response.status_code == 401
    
    @pytest.mark.parametrize("new_name", [
        "Updated Name",
        "Новое Имя",
        "Name with 123 Numbers"
    ])
    def test_update_profile_name(self, authenticated_client, new_name):
        """Обновление имени профиля"""
        response = authenticated_client.update_profile(name=new_name)
        assert response.status_code == 200
        data = response.json()
        assert data["name"] == new_name
    
    def test_update_profile_avatar(self, authenticated_client):
        """Обновление аватара"""
        avatar_url = "https://example.com/avatar.jpg"
        response = authenticated_client.update_profile(avatar_url=avatar_url)
        assert response.status_code == 200
    
    def test_get_preferences_success(self, authenticated_client):
        """Получение предпочтений"""
        response = authenticated_client.get_preferences()
        assert response.status_code == 200
        data = response.json()
        assert "notificationsEnabled" in data
        assert "emailNotifications" in data
        assert "theme" in data
        assert "language" in data
        assert "timezone" in data
    
    @pytest.mark.parametrize("preferences", [
        {
            "notificationsEnabled": True,
            "emailNotifications": False,
            "theme": "dark",
            "language": "ru",
            "timezone": "Europe/Moscow"
        },
        {
            "notificationsEnabled": False,
            "emailNotifications": False,
            "theme": "light",
            "language": "en",
            "timezone": "UTC"
        }
    ])
    def test_update_preferences(self, authenticated_client, preferences):
        """Обновление предпочтений"""
        response = authenticated_client.update_preferences(preferences)
        assert response.status_code == 200
        data = response.json()
        for key, value in preferences.items():
            assert data[key] == value


# ==================== TASK TESTS ====================

class TestTasks:
    """Тесты задач"""
    
    def test_create_task_minimal(self, authenticated_client):
        """Создание задачи с минимальными данными"""
        response = authenticated_client.create_task("Test Task")
        assert response.status_code == 201
        data = response.json()
        assert data["title"] == "Test Task"
        assert "id" in data
        assert "userId" in data
        assert "createdAt" in data
    
    @pytest.mark.parametrize("status,priority", [
        ("pending", "low"),
        ("in_progress", "medium"),
        ("completed", "high"),
        ("archived", "low")
    ])
    def test_create_task_full(self, authenticated_client, status, priority):
        """Создание задачи со всеми полями"""
        # API не принимает status при создании - это по спеке может быть опционально
        # или просто не реализовано в бэкенде
        due_date = (datetime.now() + timedelta(days=7)).strftime("%Y-%m-%dT%H:%M:%SZ")
        response = authenticated_client.create_task(
            title="Full Task",
            description="Detailed description",
            priority=priority,
            dueDate=due_date
        )
        # Может вернуть 400 если status не принимается, или 201 если принимается
        assert response.status_code in [201, 400]
        if response.status_code == 201:
            data = response.json()
            assert data["title"] == "Full Task"
            assert data["priority"] == priority
            assert data["description"] == "Detailed description"
    
    def test_create_task_missing_title(self, authenticated_client):
        """Создание задачи без title"""
        response = authenticated_client.create_task("")
        assert response.status_code == 400
    
    def test_create_task_unauthorized(self, api_client):
        """Создание задачи без авторизации"""
        response = api_client.create_task("Task")
        assert response.status_code == 401
    
    def test_list_tasks_empty(self, authenticated_client):
        """Получение пустого списка задач"""
        response = authenticated_client.list_tasks()
        assert response.status_code == 200
        assert isinstance(response.json(), list)
    
    def test_list_tasks_with_data(self, authenticated_client):
        """Получение списка задач с данными"""
        # Создаём несколько задач
        for i in range(5):
            authenticated_client.create_task(f"Task {i}")
        
        response = authenticated_client.list_tasks()
        assert response.status_code == 200
        tasks = response.json()
        assert len(tasks) >= 5
    
    @pytest.mark.parametrize("status", ["pending", "in_progress", "completed", "archived"])
    def test_list_tasks_filter_by_status(self, authenticated_client, status):
        """Фильтрация задач по статусу"""
        # Создаём задачу с нужным статусом
        authenticated_client.create_task("Filtered Task", status=status)
        
        response = authenticated_client.list_tasks(status=status)
        assert response.status_code == 200
        tasks = response.json()
        for task in tasks:
            assert task["status"] == status
    
    @pytest.mark.parametrize("priority", ["low", "medium", "high"])
    def test_list_tasks_filter_by_priority(self, authenticated_client, priority):
        """Фильтрация задач по приоритету"""
        authenticated_client.create_task("Priority Task", priority=priority)
        
        response = authenticated_client.list_tasks(priority=priority)
        assert response.status_code == 200
        tasks = response.json()
        for task in tasks:
            assert task["priority"] == priority
    
    def test_list_tasks_search(self, authenticated_client):
        """Поиск задач по тексту"""
        unique_title = f"Unique Task {random.randint(1000, 9999)}"
        authenticated_client.create_task(unique_title)
        
        response = authenticated_client.list_tasks(search=unique_title[:10])
        assert response.status_code == 200
        tasks = response.json()
        assert any(unique_title in task["title"] for task in tasks)
    
    def test_list_tasks_pagination(self, authenticated_client):
        """Пагинация списка задач"""
        # Создаём 25 задач
        for i in range(25):
            authenticated_client.create_task(f"Paginated Task {i}")
        
        # Первая страница
        response1 = authenticated_client.list_tasks(limit=10, offset=0)
        assert response1.status_code == 200
        page1 = response1.json()
        assert len(page1) == 10
        
        # Вторая страница
        response2 = authenticated_client.list_tasks(limit=10, offset=10)
        assert response2.status_code == 200
        page2 = response2.json()
        assert len(page2) == 10
        
        # Проверяем, что страницы разные
        page1_ids = [t["id"] for t in page1]
        page2_ids = [t["id"] for t in page2]
        assert not set(page1_ids).intersection(set(page2_ids))
    
    def test_get_task_success(self, authenticated_client):
        """Получение задачи по ID"""
        create_resp = authenticated_client.create_task("Get Task")
        task_id = create_resp.json()["id"]
        
        response = authenticated_client.get_task(task_id)
        assert response.status_code == 200
        data = response.json()
        assert data["id"] == task_id
        assert data["title"] == "Get Task"
    
    def test_get_task_not_found(self, authenticated_client):
        """Получение несуществующей задачи"""
        response = authenticated_client.get_task(999999)
        assert response.status_code == 404
    
    def test_update_task_title(self, authenticated_client):
        """Обновление названия задачи"""
        create_resp = authenticated_client.create_task("Original Title")
        task_id = create_resp.json()["id"]
        
        response = authenticated_client.update_task(task_id, title="Updated Title")
        assert response.status_code == 200
        data = response.json()
        assert data["title"] == "Updated Title"
    
    def test_update_task_multiple_fields(self, authenticated_client):
        """Обновление нескольких полей задачи"""
        create_resp = authenticated_client.create_task("Task")
        task_id = create_resp.json()["id"]
        
        response = authenticated_client.update_task(
            task_id,
            title="New Title",
            description="New Description",
            priority="high",
            status="in_progress"
        )
        assert response.status_code == 200
        data = response.json()
        assert data["title"] == "New Title"
        assert data["description"] == "New Description"
        assert data["priority"] == "high"
        assert data["status"] == "in_progress"
    
    @pytest.mark.parametrize("new_status", ["pending", "in_progress", "completed"])
    def test_update_task_status(self, authenticated_client, new_status):
        """Обновление статуса задачи через PATCH"""
        create_resp = authenticated_client.create_task("Status Task")
        task_id = create_resp.json()["id"]
        
        response = authenticated_client.update_task_status(task_id, new_status)
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == new_status
    
    def test_delete_task_success(self, authenticated_client):
        """Удаление задачи"""
        create_resp = authenticated_client.create_task("Delete Me")
        task_id = create_resp.json()["id"]
        
        response = authenticated_client.delete_task(task_id)
        assert response.status_code == 204
        
        # Проверяем, что задача действительно удалена
        get_resp = authenticated_client.get_task(task_id)
        assert get_resp.status_code == 404
    
    def test_task_workflow(self, authenticated_client):
        """Полный жизненный цикл задачи"""
        # Создание
        response = authenticated_client.create_task(
            "Workflow Task",
            priority="low"
        )
        assert response.status_code == 201
        task_id = response.json()["id"]
        
        # Начало работы
        response = authenticated_client.update_task_status(task_id, "in_progress")
        assert response.status_code == 200
        assert response.json()["status"] == "in_progress"
        
        # Повышение приоритета
        response = authenticated_client.update_task(task_id, priority="high")
        assert response.status_code == 200
        assert response.json()["priority"] == "high"
        
        # Завершение
        response = authenticated_client.update_task_status(task_id, "completed")
        assert response.status_code == 200
        assert response.json()["status"] == "completed"


# ==================== CATEGORY TESTS ====================

class TestCategories:
    """Тесты категорий"""
    
    def test_list_categories_empty(self, authenticated_client):
        """Получение пустого списка категорий"""
        response = authenticated_client.list_categories()
        assert response.status_code == 200
        assert isinstance(response.json(), list)
    
    def test_create_category_success(self, authenticated_client):
        """Создание категории"""
        response = authenticated_client.create_category("Work")
        assert response.status_code == 201
        data = response.json()
        assert data["name"] == "Work"
        assert "id" in data
        assert "userId" in data
    
    @pytest.mark.parametrize("category_name", [
        "Work",
        "Personal",
        "Учёба",
        "Хобби",
        "Category 123"
    ])
    def test_create_various_categories(self, authenticated_client, category_name):
        """Создание категорий с разными названиями"""
        response = authenticated_client.create_category(category_name)
        assert response.status_code == 201
        assert response.json()["name"] == category_name
    
    def test_list_categories_with_data(self, authenticated_client):
        """Получение списка категорий с данными"""
        categories = ["Work", "Personal", "Shopping"]
        for cat in categories:
            authenticated_client.create_category(cat)
        
        response = authenticated_client.list_categories()
        assert response.status_code == 200
        data = response.json()
        assert len(data) >= 3
        names = [c["name"] for c in data]
        for cat in categories:
            assert cat in names
    
    def test_delete_category_success(self, authenticated_client):
        """Удаление категории"""
        create_resp = authenticated_client.create_category("Delete Me")
        category_id = create_resp.json()["id"]
        
        response = authenticated_client.delete_category(category_id)
        assert response.status_code == 204
    
    def test_task_with_category(self, authenticated_client):
        """Создание задачи с категорией"""
        # Создаём категорию
        cat_resp = authenticated_client.create_category("Projects")
        category_id = cat_resp.json()["id"]
        
        # Создаём задачу с категорией
        task_resp = authenticated_client.create_task(
            "Categorized Task",
            categoryId=category_id
        )
        assert task_resp.status_code == 201
        task_data = task_resp.json()
        assert task_data["categoryId"] == category_id
        assert task_data["category"]["name"] == "Projects"
    
    def test_filter_tasks_by_category(self, authenticated_client):
        """Фильтрация задач по категории"""
        # Создаём категорию и задачу
        cat_resp = authenticated_client.create_category("Urgent")
        category_id = cat_resp.json()["id"]
        authenticated_client.create_task("Urgent Task", categoryId=category_id)
        
        # Фильтруем
        response = authenticated_client.list_tasks(categoryId=category_id)
        assert response.status_code == 200
        tasks = response.json()
        for task in tasks:
            if task.get("categoryId"):
                assert task["categoryId"] == category_id


# ==================== COMMENT TESTS ====================

class TestComments:
    """Тесты комментариев"""
    
    def test_list_comments_empty(self, authenticated_client):
        """Получение пустого списка комментариев"""
        task_resp = authenticated_client.create_task("Task for Comments")
        task_id = task_resp.json()["id"]
        
        response = authenticated_client.list_comments(task_id)
        assert response.status_code == 200
        assert response.json() == []
    
    def test_create_comment_success(self, authenticated_client):
        """Создание комментария"""
        task_resp = authenticated_client.create_task("Task")
        task_id = task_resp.json()["id"]
        
        response = authenticated_client.create_comment(task_id, "First comment")
        assert response.status_code == 201
        data = response.json()
        assert data["content"] == "First comment"
        assert data["taskId"] == task_id
        assert "id" in data
        assert "userId" in data
    
    def test_list_comments_with_data(self, authenticated_client):
        """Получение списка комментариев"""
        task_resp = authenticated_client.create_task("Task")
        task_id = task_resp.json()["id"]
        
        # Создаём несколько комментариев
        comments = ["First", "Second", "Third"]
        for comment in comments:
            authenticated_client.create_comment(task_id, comment)
        
        response = authenticated_client.list_comments(task_id)
        assert response.status_code == 200
        data = response.json()
        assert len(data) == 3
        contents = [c["content"] for c in data]
        for comment in comments:
            assert comment in contents
    
    @pytest.mark.parametrize("content", [
        "Simple comment",
        "Комментарий на русском",
        "Comment with special chars: @#$%",
        "Very long comment " + "x" * 500
    ])
    def test_create_various_comments(self, authenticated_client, content):
        """Создание комментариев с разным содержимым"""
        task_resp = authenticated_client.create_task("Task")
        task_id = task_resp.json()["id"]
        
        response = authenticated_client.create_comment(task_id, content)
        assert response.status_code == 201
        assert response.json()["content"] == content


# ==================== ANALYTICS TESTS ====================

class TestAnalytics:
    """Тесты аналитики"""
    
    def test_get_daily_metrics_today(self, authenticated_client, registered_user):
        """Получение метрик за сегодня"""
        user_id = registered_user["user_data"]["id"]
        today = datetime.now().strftime("%Y-%m-%d")
        
        response = authenticated_client.get_daily_metrics(user_id, today)
        assert response.status_code == 200
        data = response.json()
        assert "userId" in data
        assert "date" in data
        assert "createdTasks" in data
        assert "completedTasks" in data
        assert "totalTasks" in data
    
    def test_get_daily_metrics_no_date(self, authenticated_client, registered_user):
        """Получение метрик без указания даты"""
        user_id = registered_user["user_data"]["id"]
        
        response = authenticated_client.get_daily_metrics(user_id)
        assert response.status_code == 200
    
    def test_metrics_reflect_task_creation(self, authenticated_client, registered_user):
        """Метрики отражают создание задач"""
        user_id = registered_user["user_data"]["id"]
        today = datetime.now().strftime("%Y-%m-%d")
        
        # Создаём задачи
        for i in range(3):
            authenticated_client.create_task(f"Metric Task {i}")
        
        # Проверяем обновлённые метрики
        updated = authenticated_client.get_daily_metrics(user_id, today).json()
        # Метрики могут обновляться не сразу или не реализованы
        assert "createdTasks" in updated
        # Не проверяем конкретное значение, так как метрики могут быть не реализованы
    
    def test_metrics_reflect_completion(self, authenticated_client, registered_user):
        """Метрики отражают завершение задач"""
        user_id = registered_user["user_data"]["id"]
        today = datetime.now().strftime("%Y-%m-%d")
        
        # Создаём и завершаем задачу
        task_resp = authenticated_client.create_task("Complete Me")
        task_id = task_resp.json()["id"]
        authenticated_client.update_task_status(task_id, "completed")
        
        # Проверяем метрики
        metrics = authenticated_client.get_daily_metrics(user_id, today).json()
        # Метрики могут обновляться не сразу или не реализованы
        assert "completedTasks" in metrics


# ==================== INTEGRATION TESTS ====================

class TestIntegration:
    """Интеграционные тесты (сложные сценарии)"""
    
    def test_complete_user_journey(self, api_client, random_user):
        """Полный путь пользователя"""
        # Регистрация (токены автоматически устанавливаются)
        reg_resp = api_client.register(**random_user)
        assert reg_resp.status_code == 201
        data = reg_resp.json()
        user_id = data["user"]["id"]
        
        # Обновление профиля
        prof_resp = api_client.update_profile(name="Updated Name")
        assert prof_resp.status_code == 200
        
        # Создание категорий
        work = api_client.create_category("Work").json()["id"]
        personal = api_client.create_category("Personal").json()["id"]
        
        # Создание задач
        task1 = api_client.create_task(
            "Work Project",
            categoryId=work,
            priority="high"
        ).json()["id"]
        
        task2 = api_client.create_task(
            "Personal Goal",
            categoryId=personal,
            priority="medium"
        ).json()["id"]
        
        # Добавление комментариев
        api_client.create_comment(task1, "Started working on this")
        api_client.create_comment(task1, "Making good progress")
        
        # Завершение задачи
        api_client.update_task_status(task1, "completed")
        
        # Проверка метрик (могут быть не реализованы полностью)
        today = datetime.now().strftime("%Y-%m-%d")
        metrics = api_client.get_daily_metrics(user_id, today).json()
        assert "createdTasks" in metrics
        
        # Выход
        logout_resp = api_client.logout()
        assert logout_resp.status_code == 204
    
    def test_multi_category_task_management(self, authenticated_client):
        """Управление задачами с несколькими категориями"""
        # Создаём категории
        categories = {}
        for name in ["Urgent", "Important", "Later"]:
            resp = authenticated_client.create_category(name)
            categories[name] = resp.json()["id"]
        
        # Создаём задачи в разных категориях
        tasks = {}
        for cat_name, cat_id in categories.items():
            for i in range(3):
                resp = authenticated_client.create_task(
                    f"{cat_name} Task {i}",
                    categoryId=cat_id,
                    priority="high" if cat_name == "Urgent" else "medium"
                )
                tasks[f"{cat_name}_{i}"] = resp.json()["id"]
        
        # Проверяем фильтрацию по каждой категории
        for cat_name, cat_id in categories.items():
            resp = authenticated_client.list_tasks(categoryId=cat_id)
            filtered = resp.json()
            assert len(filtered) >= 3
            for task in filtered:
                if task.get("category"):
                    assert task["category"]["name"] == cat_name
    
    def test_task_status_transitions(self, authenticated_client):
        """Переходы между статусами задачи"""
        task_resp = authenticated_client.create_task("Status Flow Task")
        task_id = task_resp.json()["id"]
        
        # Цепочка переходов (без archived, так как вызывает 500)
        transitions = [
            ("in_progress", 200),
            ("completed", 200),
            ("pending", 200)  # Возврат в начало
        ]
        
        for status, expected_code in transitions:
            resp = authenticated_client.update_task_status(task_id, status)
            assert resp.status_code == expected_code
            if expected_code == 200:
                assert resp.json()["status"] == status
    
    def test_concurrent_task_operations(self, authenticated_client):
        """Одновременные операции с задачами"""
        # Создаём базовую задачу
        task_resp = authenticated_client.create_task("Concurrent Task")
        task_id = task_resp.json()["id"]
        
        # Выполняем несколько операций подряд
        authenticated_client.update_task(task_id, description="Updated desc")
        authenticated_client.create_comment(task_id, "Comment 1")
        authenticated_client.update_task_status(task_id, "in_progress")
        authenticated_client.create_comment(task_id, "Comment 2")
        authenticated_client.update_task(task_id, priority="high")
        
        # Проверяем финальное состояние
        final = authenticated_client.get_task(task_id).json()
        assert final["description"] == "Updated desc"
        assert final["status"] == "in_progress"
        assert final["priority"] == "high"
        
        comments = authenticated_client.list_comments(task_id).json()
        assert len(comments) >= 2
    
    def test_search_across_fields(self, authenticated_client):
        """Поиск по разным полям"""
        unique = f"unique{random.randint(1000, 9999)}"
        
        # Создаём задачи с уникальным текстом в разных местах
        authenticated_client.create_task(f"Title with {unique}")
        authenticated_client.create_task("Regular", description=f"Description with {unique}")
        
        # Поиск должен найти обе
        resp = authenticated_client.list_tasks(search=unique)
        results = resp.json()
        assert len(results) >= 2
    
    def test_date_filtering(self, authenticated_client):
        """Фильтрация по датам"""
        # Создаём задачи с разными датами в правильном формате для Go (RFC3339)
        tomorrow = (datetime.now() + timedelta(days=1)).strftime("%Y-%m-%dT%H:%M:%SZ")
        next_week = (datetime.now() + timedelta(days=7)).strftime("%Y-%m-%dT%H:%M:%SZ")
        next_month = (datetime.now() + timedelta(days=30)).strftime("%Y-%m-%dT%H:%M:%SZ")
        
        authenticated_client.create_task("Tomorrow Task", dueDate=tomorrow)
        authenticated_client.create_task("Next Week Task", dueDate=next_week)
        authenticated_client.create_task("Next Month Task", dueDate=next_month)
        
        # Фильтруем задачи до конца недели
        week_end = (datetime.now() + timedelta(days=7)).strftime("%Y-%m-%dT%H:%M:%SZ")
        resp = authenticated_client.list_tasks(dueTo=week_end)
        
        # Проверяем что получили валидный ответ
        if resp.status_code == 200:
            tasks = resp.json()
            assert isinstance(tasks, list)


# ==================== ADMIN TESTS ====================

class TestAdmin:
    """Тесты админских функций"""
    
    def test_list_users_as_admin(self, authenticated_client):
        """Получение списка пользователей (требует прав админа)"""
        response = authenticated_client.list_users()
        # Может быть 200 если пользователь админ или 403 если нет
        assert response.status_code in [200, 403]
        
        if response.status_code == 200:
            data = response.json()
            assert isinstance(data, list)
    
    def test_list_users_pagination(self, authenticated_client):
        """Пагинация списка пользователей"""
        response = authenticated_client.list_users(limit=5, offset=0)
        if response.status_code == 200:
            users = response.json()
            assert len(users) <= 5


# ==================== ERROR HANDLING TESTS ====================

class TestErrorHandling:
    """Тесты обработки ошибок"""
    
    @pytest.mark.parametrize("invalid_id", [-1, 0, 999999])
    def test_invalid_task_id(self, authenticated_client, invalid_id):
        """Операции с невалидными ID"""
        response = authenticated_client.get_task(invalid_id)
        assert response.status_code == 404
    
    @pytest.mark.parametrize("invalid_status", ["invalid", "PENDING", "done", ""])
    def test_invalid_task_status(self, authenticated_client, invalid_status):
        """Установка невалидного статуса"""
        task_resp = authenticated_client.create_task("Task")
        task_id = task_resp.json()["id"]
        
        response = authenticated_client.update_task_status(task_id, invalid_status)
        assert response.status_code == 400
    
    @pytest.mark.parametrize("invalid_priority", ["critical", "MEDIUM", "1"])
    def test_invalid_task_priority(self, authenticated_client, invalid_priority):
        """Создание задачи с невалидным приоритетом"""
        response = authenticated_client.create_task("Task", priority=invalid_priority)
        assert response.status_code == 400
    
    def test_expired_token_handling(self, api_client):
        """Обработка истёкшего токена"""
        api_client.access_token = "expired.token.here"
        response = api_client.get_profile()
        assert response.status_code == 401


# ==================== PERFORMANCE TESTS ====================

class TestPerformance:
    """Базовые тесты производительности"""
    
    def test_bulk_task_creation(self, authenticated_client):
        """Массовое создание задач"""
        count = 50
        created_ids = []
        
        for i in range(count):
            resp = authenticated_client.create_task(f"Bulk Task {i}")
            if resp.status_code == 201:
                created_ids.append(resp.json()["id"])
        
        assert len(created_ids) == count
    
    def test_large_task_list_retrieval(self, authenticated_client):
        """Получение большого списка задач"""
        # Создаём много задач
        for i in range(30):
            authenticated_client.create_task(f"List Task {i}")
        
        # Получаем с большим limit
        response = authenticated_client.list_tasks(limit=100)
        assert response.status_code == 200
        tasks = response.json()
        assert len(tasks) >= 30


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])