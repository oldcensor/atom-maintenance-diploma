import time
import random
from threading import Lock
from dataclasses import dataclass, field
from datetime import datetime, timezone

REFERENCE_DATE_EPOCH = 1704067200  # 2024-01-01 UTC

UNITS = {
    "operating_hours": "м/ч",
    "cycles": "цикл",
    "days": "дн",
}


@dataclass
class EquipmentItem:
    equipment_id: int
    name: str
    meter_type: str  # operating_hours | cycles | days

    _base_value: float = field(init=False, default=0.0)
    _start_time: float = field(init=False, default=0.0)
    _start_monotonic: float = field(init=False, default=0.0)

    _rate: float = field(init=False, default=0.0)  # скорость роста моточасов

    _cycle_count: float = field(init=False, default=0.0)
    _last_cycle_tick: float = field(init=False, default=0.0)

    _last_reported_value: float | None = field(init=False, default=None)
    _lock: Lock = field(init=False, default_factory=Lock)

    def __post_init__(self):
        self._start_time = time.time()
        self._start_monotonic = time.monotonic()

        if self.meter_type == "operating_hours":
            self._base_value = random.uniform(10_000, 50_000)

            # скорость фиксируется один раз
            self._rate = random.uniform(22.0, 24.0)

        elif self.meter_type == "cycles":
            self._base_value = random.uniform(500, 5_000)
            self._cycle_count = self._base_value
            self._last_cycle_tick = self._start_monotonic

        elif self.meter_type == "days":
            self._base_value = 0.0

    def current_value(self) -> float:
        with self._lock:
            raw_value = self._calculate_value()
            return self._strictly_increasing(raw_value)

    def _calculate_value(self) -> float:
        now = time.time()
        monotonic_now = time.monotonic()

        if self.meter_type == "operating_hours":
            elapsed_hours = (monotonic_now - self._start_monotonic) / 3600.0
            value = self._base_value + elapsed_hours * self._rate
            return round(value, 2)

        elif self.meter_type == "cycles":
            elapsed = monotonic_now - self._last_cycle_tick
            ticks = int(elapsed / 60)

            if ticks > 0:
                for _ in range(ticks):
                    self._cycle_count += random.randint(1, 5)

                self._last_cycle_tick += ticks * 60

            return round(self._cycle_count, 0)

        elif self.meter_type == "days":
            return round((now - REFERENCE_DATE_EPOCH) / 86400.0, 2)

        return 0.0

    def _strictly_increasing(self, value: float) -> float:
        if self._last_reported_value is None:
            self._last_reported_value = value
            return value

        step = 1.0 if self.meter_type == "cycles" else 0.01
        if value <= self._last_reported_value:
            value = self._last_reported_value + step

        if self.meter_type == "cycles":
            value = round(value, 0)
        else:
            value = round(value, 2)

        self._last_reported_value = value
        return value

    def to_dict(self) -> dict:
        return {
            "equipment_id": self.equipment_id,
            "name": self.name,
            "meter_type": self.meter_type,
            "current_value": self.current_value(),
            "unit": UNITS.get(self.meter_type, ""),
            "timestamp": datetime.now(timezone.utc).isoformat(),
        }

    def history_points(self, n: int = 60, interval_sec: int = 300) -> list[dict]:
        """
        Генерирует n исторических точек назад от текущего момента.
        Показывает три фазы:
          - 0..30%  → линейный рост (нормальная работа)
          - 30..55% → ускоренный рост (повышенная нагрузка)
          - 55..70% → плато (простой / в обслуживании)
          - 70..100%→ возобновление линейного роста
        """
        now = time.time()
        current = self.current_value()
        unit = UNITS.get(self.meter_type, "")

        # Суммарный "прирост" за историческое окно (назад во времени)
        total_seconds = n * interval_sec
        if self.meter_type == "operating_hours":
            total_growth = (total_seconds / 3600.0) * self._rate * 1.15
        elif self.meter_type == "cycles":
            total_growth = n * 3.0  # примерный прирост за всё окно
        else:  # days
            total_growth = total_seconds / 86400.0

        points = []
        for i in range(n):
            age = (n - 1 - i) * interval_sec          # секунд назад от сейчас
            t = now - age
            frac = i / (n - 1) if n > 1 else 1.0      # 0.0 → самая старая, 1.0 → текущая

            # Накопленный прирост до этой точки (площадь под кривой)
            if frac < 0.30:
                # линейный рост: равномерно набираем 25% от total_growth
                cumulative = total_growth * 0.25 * (frac / 0.30)
            elif frac < 0.55:
                # ускоренный: добавляем ещё 45%
                cumulative = total_growth * 0.25 + total_growth * 0.45 * ((frac - 0.30) / 0.25)
            elif frac < 0.70:
                # плато: значение почти не растёт (добавляем 1%)
                cumulative = total_growth * 0.70 + total_growth * 0.01 * ((frac - 0.55) / 0.15)
            else:
                # возобновление: набираем оставшиеся 29%
                cumulative = total_growth * 0.71 + total_growth * 0.29 * ((frac - 0.70) / 0.30)

            value = round(current - total_growth + cumulative, 2)
            ts = datetime.fromtimestamp(t, tz=timezone.utc).isoformat()
            points.append({"timestamp": ts, "value": value, "unit": unit})

        return points


EQUIPMENT_LIST = [
    EquipmentItem(equipment_id=1, name="ГЦН-1", meter_type="operating_hours"),
    EquipmentItem(equipment_id=2, name="ГЦН-2", meter_type="operating_hours"),
    EquipmentItem(equipment_id=3, name="Кран-регулятор КР-7", meter_type="cycles"),
    EquipmentItem(equipment_id=4, name="Аварийный дизель-генератор", meter_type="operating_hours"),
    EquipmentItem(equipment_id=5, name="Теплообменник ТО-3", meter_type="days"),
]

EQUIPMENT_MAP = {e.equipment_id: e for e in EQUIPMENT_LIST}
