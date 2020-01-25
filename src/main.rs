use irc::client::prelude::*;

type Result<T> = std::result::Result<T, irc::error::IrcError>;

fn main() -> Result<()> {
    println!("dungeonbot 0.1");
    println!();

    let client = IrcClient::new("./conf.toml")?;
    let mut reacc = IrcReactor::new()?;
    client.identify()?;

    Ok(())
}
