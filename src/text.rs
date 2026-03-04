use std::path::Path;

pub fn dir_or_dot(path: &str) -> String {
    let p = Path::new(path);
    p.parent()
        .filter(|parent| !parent.as_os_str().is_empty())
        .map(|parent| parent.to_string_lossy().to_string())
        .unwrap_or_else(|| ".".to_string())
}

pub fn tail(s: &str, n: usize) -> String {
    let len = s.chars().count();
    if n >= len {
        return s.to_string();
    }
    s.chars().skip(len - n).collect()
}

pub fn preview(s: &str, n: usize) -> String {
    let single_line = s.replace('\n', " ");
    if single_line.chars().count() <= n {
        return single_line;
    }
    format!("{}...", single_line.chars().take(n).collect::<String>())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_dir_or_dot() {
        let cases = [
            ("file.txt", "."),
            ("./file.txt", "."),
            ("a/b/c.txt", "a/b"),
            ("/tmp/foo/bar", "/tmp/foo"),
            ("/single", "/"),
            (".", "."),
            ("", "."),
        ];

        for (input, expected) in cases {
            assert_eq!(dir_or_dot(input), expected);
        }
    }

    #[test]
    fn test_tail() {
        let s = "abcdefghijklmnopqrstuvwxyz";
        assert_eq!(tail(s, 5), "vwxyz");
        assert_eq!(tail(s, 100), s);
    }

    #[test]
    fn test_preview() {
        let s = "line1\nline2\nline3";
        assert!(!preview(s, 100).is_empty());
        assert_ne!(preview(s, 100), s);
        assert_eq!(preview("aaaaaaaaaa", 5), "aaaaa...");
    }
}
