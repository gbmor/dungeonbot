use irc::client::prelude::*;

type Result<T> = std::result::Result<T, irc::error::IrcError>;

#[cfg_attr(tarpaulin, skip)]
fn main() -> Result<()> {
    let channels: Vec<String> = vec!["#d&d".to_string()];
    const NICK: &str = "dungeonbot";
    const SERVER: &str = "irc.tilde.chat";
    const PORT: u16 = 6697;
    const USE_SSL: bool = true;

    println!("dungeonbot 0.1");
    println!();

    let config = Config {
        nickname: Some(NICK.to_string()),
        server: Some(SERVER.to_string()),
        port: Some(PORT),
        use_ssl: Some(USE_SSL),
        channels: Some(channels),
        ..Config::default()
    };

    let client = IrcClient::from_config(config)?;
    client.identify()?;

    client.for_each_incoming(|ircmsg| {
        if let Command::PRIVMSG(chan, msg) = ircmsg.command {
            eprintln!("{}", msg);
        }
    })?;

    Ok(())
}
