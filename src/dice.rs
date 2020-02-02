use rand::prelude::*;

fn roll(msg: &str) -> String {
    unimplemented!();
    let mut rng = rand::thread_rng();

    let chars = msg.chars().map(|c| c.to_string()).collect::<Vec<String>>();
    let dice_type = String::from("6"); // format!("{}", chars[2..]);

    let upper_bound: u32 = if let Ok(val) = dice_type.parse::<u32>() {
        val + 1
    } else {
        21
    };

    let num_dice: u32 = if let Ok(val) = chars[0].parse() {
        val
    } else {
        1
    };

    let mut output: String = String::new();
    (1..=num_dice).for_each(|_| {
        let roll = rng.gen_range(1, upper_bound);
        output.push_str(&format!(" {}", roll));
    });

    output
}
