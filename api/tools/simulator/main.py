from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from rich.console import Console
from rich.table import Table
from rich import box
from equipment import EQUIPMENT_LIST, EQUIPMENT_MAP

app = FastAPI(title="TOIR Equipment Simulator", version="1.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["GET"],
    allow_headers=["*"],
)


def _print_startup_table():
    console = Console()
    console.print()
    console.print("[bold cyan]⚙  TOIR Equipment Simulator[/bold cyan]", justify="center")
    console.print()

    table = Table(box=box.ROUNDED, show_header=True, header_style="bold white on dark_blue")
    table.add_column("ID", style="bold yellow", justify="center", width=6)
    table.add_column("Название", style="white", min_width=32)
    table.add_column("Счётчик", style="cyan", width=18)

    meter_label = {
        "operating_hours": "Моточасы",
        "cycles": "Циклы",
        "days": "Дни (кал.)",
    }

    for e in EQUIPMENT_LIST:
        table.add_row(
            str(e.equipment_id),
            e.name,
            meter_label.get(e.meter_type, e.meter_type),
        )

    console.print(table, justify="center")
    console.print()
    console.print(
        "[bold green]Создайте оборудование в API командой:[/bold green] "
        "[bold white]make seed-equipment[/bold white]"
    )
    console.print(
        "[dim]Телеметрия доступна на[/dim] [bold]http://localhost:8090/api/v1/telemetry[/bold]"
    )
    console.print()


_print_startup_table()


@app.get("/health")
def health():
    return {"status": "ok"}


@app.get("/api/v1/telemetry")
def get_all():
    return [e.to_dict() for e in EQUIPMENT_LIST]


@app.get("/api/v1/telemetry/{equipment_id}")
def get_by_id(equipment_id: int):
    item = EQUIPMENT_MAP.get(equipment_id)
    if item is None:
        raise HTTPException(status_code=404, detail="equipment not found")
    return item.to_dict()


@app.get("/api/v1/telemetry/{equipment_id}/history")
def get_history(equipment_id: int, n: int = 60, interval: int = 300):
    item = EQUIPMENT_MAP.get(equipment_id)
    if item is None:
        raise HTTPException(status_code=404, detail="equipment not found")
    return item.history_points(n=min(n, 200), interval_sec=interval)
