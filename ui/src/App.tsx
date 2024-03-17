import { For, createSignal } from "solid-js";

const ws = new WebSocket("ws://127.0.0.1:3000/ws");

type UpdatedRow = `${number};${number};${string}`;
type Row = { x: number; y: number; c: string };

const colors = [
	"#E74C3C",
	"#E67E22",
	"#F1C40F",
	"#2ECC71",
	"#3498DB",
	"#9B59B6",
	"#FFFFFF",
	"#2C3E50",
];

function App() {
	const [selectedColor, setSelectedColor] = createSignal<string>(colors[0]);
	const [rows, setRows] = createSignal<Row[][]>([]);

	ws.addEventListener("message", (event) => {
		const data: Row[] | UpdatedRow = JSON.parse(event.data);

		if (Array.isArray(data)) {
			const grid: Row[][] = [];
			for (const elem of data) {
				grid[elem.x] = grid[elem.x] || [];
				grid[elem.x][elem.y] = elem;
			}
			setRows(grid);
		} else {
			setRows((prev) => {
				const tokens = data.split(";");
				const x = parseInt(tokens[0]);
				const y = parseInt(tokens[1]);

				prev[x][y] = {
					x,
					y,
					c: tokens[2],
				};

				prev[x] = [...prev[x]];
				return [...prev];
			});
		}
	});

	const handleClick = (row: Row) => {
		if (row.c === selectedColor()) {
			return;
		}

		ws.send(
			JSON.stringify({
				event: "click",
				contents: `${row.x};${row.y};${selectedColor()}`,
			}),
		);
	};

	return (
		<section
			style={{
				"margin-left": "auto",
				"margin-right": "auto",
				width: "100%",
				"max-width": "fit-content",
			}}
		>
			<div style={{ display: "flex", "padding-bottom": "2rem" }}>
				<For each={colors}>
					{(color) => (
						<div
							style={{
								width: "20px",
								height: "20px",
								"background-color": color,
							}}
							onclick={() => setSelectedColor(color)}
						></div>
					)}
				</For>
			</div>

			<div
				style={{
					display: "flex",
					"flex-direction": "column",
					border: "1px solid #444",
				}}
			>
				<For each={rows()}>
					{(col) => (
						<div style={{ display: "flex" }}>
							<For each={col}>
								{(elem) => (
									<div
										style={{
											width: "20px",
											height: "20px",
											"background-color": elem.c,
										}}
										onclick={() => {
											handleClick(elem);
										}}
									></div>
								)}
							</For>
						</div>
					)}
				</For>
			</div>
		</section>
	);
}

export default App;
