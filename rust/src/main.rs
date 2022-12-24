use ginger::graph::Graph;

fn main() {

    let g = Graph::new().
        with("foo", Graph::from_value(1, "bar"));

    let g1 = g.
        with("foo", Graph::from_value(1, "bar")).
        with("foo", Graph::from_value(2, "baz"));

    let g2 = g.
        with("bar", Graph::from_tuple(100, vec![
            Graph::from_value(20, "a"),
            Graph::from_value(40, "b"),
            Graph::from_value(60, "c"),
        ]));

    dbg!(g1 == g2);
    dbg!(&g1);
    dbg!(&g2);
}
